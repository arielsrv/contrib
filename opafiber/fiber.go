package opafiber

import (
	"context"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/open-policy-agent/opa/v1/rego"
)

type InputCreationFunc func(c *fiber.Ctx) (map[string]interface{}, error)

type Config struct {
	RegoPolicy            io.Reader
	RegoQuery             string
	IncludeHeaders        []string
	IncludeQueryString    bool
	DeniedStatusCode      int
	DeniedResponseMessage string
	InputCreationMethod   InputCreationFunc
}

func New(cfg Config) fiber.Handler {
	err := cfg.fillAndValidate()
	if err != nil {
		panic(err)
	}
	readedBytes, err := io.ReadAll(cfg.RegoPolicy)
	if err != nil {
		panic(fmt.Sprint("could not read rego policy %w", err))
	}
	query, err := rego.New(
		rego.Query(cfg.RegoQuery),
		rego.Module("policy.rego", utils.UnsafeString(readedBytes)),
	).PrepareForEval(context.Background())
	if err != nil {
		panic(fmt.Sprint("opa policy error: %w", err))
	}
	return func(c *fiber.Ctx) error {
		input, err := cfg.InputCreationMethod(c)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusInternalServerError)
			c.Response().SetBodyString(fmt.Sprintf("Error creating input: %s", err))
			return err
		}
		if cfg.IncludeQueryString {
			queryStringData := make(map[string][]string)
			for k, v := range c.Request().URI().QueryArgs().All() {
				queryStringData[utils.UnsafeString(k)] = append(queryStringData[utils.UnsafeString(k)], utils.UnsafeString(v))
			}
			input["query"] = queryStringData
		}
		if len(cfg.IncludeHeaders) > 0 {
			headers := make(map[string]string)
			for _, header := range cfg.IncludeHeaders {
				headers[header] = c.Get(header)
			}
			input["headers"] = headers
		}
		res, err := query.Eval(context.Background(), rego.EvalInput(input))
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusInternalServerError)
			c.Response().SetBodyString(fmt.Sprintf("Error evaluating opa policy: %s", err))
			return err
		}
		allowed := false
		if len(res) > 0 {
			if b, ok := res[0].Expressions[0].Value.(bool); ok {
				allowed = b
			}
		}
		if !allowed {
			c.Response().SetStatusCode(cfg.DeniedStatusCode)
			c.Response().SetBodyString(cfg.DeniedResponseMessage)
			return nil
		}
		return c.Next()
	}
}

func (c *Config) fillAndValidate() error {
	if c.RegoQuery == "" {
		return fmt.Errorf("rego query can not be empty")
	}

	if c.DeniedStatusCode == 0 {
		c.DeniedStatusCode = fiber.StatusBadRequest
	}
	if c.DeniedResponseMessage == "" {
		c.DeniedResponseMessage = fiber.ErrBadRequest.Error()
	}
	if c.IncludeHeaders == nil {
		c.IncludeHeaders = []string{}
	}
	if c.InputCreationMethod == nil {
		c.InputCreationMethod = defaultInput
	}
	return nil
}

func defaultInput(ctx *fiber.Ctx) (map[string]interface{}, error) {
	input := map[string]interface{}{
		"method": ctx.Method(),
		"path":   ctx.Path(),
	}
	return input, nil
}
