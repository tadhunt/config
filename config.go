package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"reflect"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SecretMgr = "secretmgr:"
)

func SecretPath(project string, name string, version string) string {
	return fmt.Sprintf("secretmgr:projects/%s/secrets/%s/versions/%s", project, name, version)
}

func Parse(ctx context.Context, src string, config any) error {
	var err error
	var data []byte
	if strings.HasPrefix(src, SecretMgr) {
		path := src[len(SecretMgr):]
		data, err = loadSecret(ctx, path)
	} else {
		data, err = os.ReadFile(src)
	}
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

func loadSecret(ctx context.Context, path string) ([]byte, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load %s: %v", path, err)
	}

	data := make([]byte, base64.StdEncoding.DecodedLen(len(result.Payload.Data)))

	payload := result.Payload.Data
	if payload[0] == '"' {
		payload = payload[1:]
	}
	if payload[len(payload)-1] == '"' {
		payload = payload[:len(payload)-1]
	}

        n, err := base64.StdEncoding.Decode(data, payload)
        if err != nil {
                return result.Payload.Data, err
        }
	data = data[:n]

	return data, nil
}

func Serialize(config any) ([]byte, error) {
	raw, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return nil, err
	}

	return raw, err
}

func Dump(config any, dst string) error {
	raw, err := Serialize(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, raw, 0600)
	if err != nil {
		return err
	}

	return nil
}

func SaveSecret(ctx context.Context, project string, name string, cfg any) (string, error) {
	data, err := Serialize(cfg)
	if err != nil {
		return "", err
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	creq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", project),
		SecretId: name,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	_, err = client.CreateSecret(ctx, creq)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.AlreadyExists {
			return "", fmt.Errorf("create %s: %v", name, err)
		}
		// already exists, add a new version
	}

	areq := &secretmanagerpb.AddSecretVersionRequest{
		Parent:  fmt.Sprintf("projects/%s/secrets/%s", project, name),
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}

	version, err := client.AddSecretVersion(ctx, areq)
	if err != nil {
		return "", fmt.Errorf("add version %s: %v", name, err)
	}

	return version.Name, nil
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
