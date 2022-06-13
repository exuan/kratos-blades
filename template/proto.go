package template

import (
	"bytes"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/exuan/kratos-blades/pkg/strs"
)

const protoTemplate = `
syntax = "proto3";

package {{.Service}}.{{.Version}};
import "google/api/annotations.proto";
import "validate/validate.proto";
import "google/protobuf/empty.proto";
import "protoc-gen-openapiv2/options/annotations.proto";
{{- if .IsGenEnt }} 
import "options/ent/opts.proto";
{{- end }}

option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "{{.Service}}",
    description: "{{.Service}}",
  }
  consumes: "application/json";
  produces: "application/json";
  security_definitions:    {
    security: {
      key: "{{.ApiKey}}";
      value: {
        type: TYPE_API_KEY;
        in: IN_HEADER;
        name: "your token";
      };
    }
  };
};

service {{.Service}} { {{range .Methods}}

 // {{ .Name }}

  rpc {{ .Name }}s ({{ .Name }}sRequest) returns ({{ .Name }}sReply) {
    option (google.api.http) = {
      get: "{{.Uri}}s"
    };
    
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "{{ .Name }}List";
      tags: ["{{ .Name }}"];
      {{- if $.IsAuth }}
      security: {
        security_requirement: {
          key: "{{$.ApiKey}}";
          value: {};
        }
      }
     {{- end }}
    };
  }

  rpc {{ .Name }} (IdRequest) returns ({{ .Name }}Reply) {
    option (google.api.http) = {
      get: "{{.Uri}}"
    };
   
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "{{ .Name }}";
      tags: ["{{ .Name }}"];
      {{- if $.IsAuth }}
      security: {
        security_requirement: {
          key: "{{$.ApiKey}}";
          value: {};
        }
      }
     {{- end }}
    };
  }

  rpc Save{{ .Name }} (Save{{ .Name }}Request) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post: "{{.Uri}}/save",
      body: "*"
    };

    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "{{ .Name }}Save";
      tags: ["{{ .Name }}"];
      {{- if $.IsAuth }}
      security: {
        security_requirement: {
          key: "{{$.ApiKey}}";
          value: {};
        }
      }
     {{- end }}
    };
  }

  rpc Delete{{ .Name }} (DeleteIdRequest) returns(google.protobuf.Empty) {
    option (google.api.http) = {
      post: "{{.Uri}}/delete",
      body: "*"
    };
    
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "{{ .Name }}Delete";
      tags: ["{{ .Name }}"];
      {{- if $.IsAuth }}
      security: {
        security_requirement: {
          key: "{{$.ApiKey}}";
          value: {};
        }
      }
     {{- end }}
    };
  }{{end}}
}

message DeleteIdRequest {
  int64 id = 1;
}

message IdRequest {
  int64 id = 1;
}
{{range .Methods}} 
//{{ .Name }}

message {{ .Name }}sRequest {
  int64 page = 1;
  int64 page_size = 2;
  int64 is_all = 3;
}

message {{ .Name }}sReply {
  int64 total = 1;
  int64 page = 2;
  int64 page_size = 3;
  repeated {{ .Name }}Reply items = 4;
}

message {{ .Name }}Reply {
{{- if $.IsGenEnt }} 
  option (ent.schema) = {gen: true, name: "{{ .Name }}"}; 
{{- end }}
}

message Save{{ .Name }}Request {
}
{{end}}
`

// Proto is a proto generator.
type Proto struct {
	Methods  []Method
	Path     string
	ApiKey   string
	Service  string
	Uri      string
	Version  string
	IsAuth   bool
	IsGenEnt bool
}

var ProtoCmd = &cli.Command{
	Name:        "proto",
	Description: "Generate proto file",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "methods", Aliases: []string{"m"}, Usage: "service method name, like \"User,Login\" (,)separated"},
		&cli.StringFlag{Name: "service", Aliases: []string{"s"}, DefaultText: "service", Value: "service", Usage: "service name, like backend"},
		&cli.StringFlag{Name: "api-key", Aliases: []string{"k"}, DefaultText: "api_key", Value: "api_key", Usage: "http auth key name"},
		&cli.StringFlag{Name: "path", Aliases: []string{"p"}, DefaultText: "./", Usage: "generate proto file path"},
		&cli.StringFlag{Name: "version", Aliases: []string{"v"}, DefaultText: "v1", Value: "v1"},
		&cli.BoolFlag{Name: "auth", Aliases: []string{"a"}, Usage: "generate api auth security option"},
		&cli.BoolFlag{Name: "gen-ent", Aliases: []string{"e"}, Usage: "generate ent option"},
	},
	Action: protoRun,
}

func protoRun(ctx *cli.Context) error {
	ms := strings.Split(ctx.String("methods"), ",")
	service := ctx.String("service")
	version := ctx.String("version")

	methods := make([]Method, 0)

	for _, method := range ms {
		method = strs.GoCamelCase(strings.TrimSpace(method))
		if method == "" {
			continue
		}
		methods = append(methods, Method{
			Name: method,
			Uri:  strs.DirCase(strs.GoCamelCase(version) + method),
		})
	}

	p := &Proto{
		Methods:  methods,
		ApiKey:   ctx.String("api-key"),
		IsAuth:   ctx.Bool("auth"),
		Path:     ctx.String("path"),
		Version:  version,
		Service:  strs.GoCamelCase(service),
		Uri:      strs.DirCase(strs.GoCamelCase(version) + service),
		IsGenEnt: ctx.Bool("gen-ent"),
	}

	return p.Generate()
}

// Generate generate a proto template.
func (p *Proto) Generate() error {
	body, err := p.execute()
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	to := path.Join(wd, p.Path)
	if _, err := os.Stat(to); os.IsNotExist(err) {
		if err := os.MkdirAll(to, 0o700); err != nil {
			return err
		}
	}
	name := path.Join(to, p.Service+".proto")
	//if _, err := os.Stat(name); !os.IsNotExist(err) {
	//	return fmt.Errorf("%s already exists", p.Name)
	//}
	return os.WriteFile(name, body, 0o644)
}

func (p *Proto) execute() ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("proto").Parse(strings.TrimSpace(protoTemplate))
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
