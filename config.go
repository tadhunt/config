package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"reflect"
)

func Parse(src string, config any) error {
	if strings.HasPrefix(src, "secretmgr:") {
		return fmt.Errorf("secretmgr unimplemented")
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	sdata := string(data)

	sdata = os.Expand(sdata, RunOrExpandEnv)

	err = json.Unmarshal([]byte(sdata), config)
	if err != nil {
		return err
	}

	return unescape(config)
}

func unescape(cfg any) error {
	dt := reflect.TypeOf(cfg)
	if dt.Kind() != reflect.Pointer {
		return fmt.Errorf("expected pointer got %T", cfg)
	}

	dv := reflect.ValueOf(cfg).Elem()
	dt = dv.Type()

	if dt.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct got %T", cfg)
	}

	for i := 0; i < dt.NumField(); i++ {
		//ft := dt.Field(i)
		fv := dv.Field(i)
		field := fv.Addr().Interface()

		switch v := field.(type) {
		default:
			if fv.Type().Kind() == reflect.Pointer && fv.Elem().Type().Kind() == reflect.Struct {
				err := unescape(field)
				if err == nil {
					return err
				}
			}
			return fmt.Errorf("unhandled type %T", v)
		case *string:
			*v = JSONUnEscape(*v)
		}
	}

	return nil
}

func RunOrExpandEnv(src string) string {
	if !strings.HasPrefix(src, "shell ") {
		return os.Getenv(src)
	}

	cmd := src[6:]
	if cmd == "" {
		return src + ": missing args"
	}

	raw, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("%s: %v", src, err)
	}

	out := string(raw)
	out, _ = strings.CutSuffix(out, "\n")
	return JSONEscape(out)
}

func JSONEscape(src string) string {
	var sb strings.Builder
	for _, r := range src {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '/': // Allowed but not required to be escaped. Escaping for consistency.
			sb.WriteString(`\/`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if r <= 0x1F {
				sb.WriteString(fmt.Sprintf("\\u%X", r))
			} else {
				sb.WriteRune(r)
			}
                }
        }
        return sb.String()
}

func JSONUnEscape(src string) string {
	var result string

	err := json.Unmarshal([]byte(src), &result)
	if err != nil {
		return src
	}

	return result
}
