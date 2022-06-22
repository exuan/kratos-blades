package template

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/emicklei/proto"
	"github.com/urfave/cli/v2"
)

const serviceTemplate = `
{{- /* delete empty line */ -}}
package service

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}
	pb "{{ .Package }}"
	"{{ .Biz }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	
	"github.com/google/wire"
	"github.com/go-kratos/kratos/v2/log"
)

type {{ .Service }} struct {
	pb.Unimplemented{{ .Service }}Server
	
	biz  *biz.Backend
	log  *log.Helper
}

var ProviderSet = wire.NewSet(New{{ .Service }})

func New{{ .Service }}(biz *biz.Backend, logger log.Logger) *{{ .Service }} {
	return &{{ .Service }}{
		biz:  biz,
		log:  log.NewHelper(logger),
	}
}
{{- $s1 := "google.protobuf.Empty" }} {{ range .Methods }} {{- if eq .Type 1 }}

func (s *{{ .Service }}) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return s.biz.{{ .Name }}(ctx, req)
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}
{{- else if eq .Type 3 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&pb.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}
{{- else if eq .Type 4 }}
func (s *{{ .Service }}Service) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}, conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}
{{- end }}
{{- end }}
`

var ServiceCmd = &cli.Command{
	Name: "service",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "proto", Aliases: []string{"p"}, Usage: "proto file", Required: true},
		&cli.StringFlag{Name: "target_dir", Aliases: []string{"t"}, DefaultText: "internal/service", Value: "internal/service"},
	},
	Action: serviceRun,
}

func serviceRun(ctx *cli.Context) error {
	reader, err := os.Open(ctx.String("proto"))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return err
	}

	var (
		pkg, biz string
		res      []*Service
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				p := strings.Split(o.Constant.Source, ";")
				pkg = p[0]

				biz = strings.Replace(pkg, "/api/", "/app/", 1)
				if len(p) > 1 && len(p[1]) > 0 {
					biz = strings.Replace(biz, "/"+p[1], "/internal/biz", 1)
				}

			}
		}),

		proto.WithService(func(s *proto.Service) {
			cs := &Service{
				Biz:     biz,
				Package: pkg,
				Service: s.Name,
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if !ok {
					continue
				}
				cs.Methods = append(cs.Methods, &Method{
					Service: s.Name, Name: r.Name, Request: r.RequestType,
					Reply: r.ReturnsType, Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
				})
			}
			res = append(res, cs)
		}),
	)

	targetDir := ctx.String("target_dir")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exsit\n", targetDir)
		return err
	}
	for _, s := range res {
		to := path.Join(targetDir, strings.ToLower(s.Service)+".go")
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			//fmt.Fprintf(os.Stderr, "%s already exists: %s\n", s.Service, to)
			//continue
		}
		b, err := s.execute()
		if err != nil {
			return err
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			return err
		}
		fmt.Println(to)
	}

	return nil
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}

type MethodType uint8

const (
	unaryType          MethodType = 1
	twoWayStreamsType  MethodType = 2
	requestStreamsType MethodType = 3
	returnsStreamsType MethodType = 4
)

// Service is a proto service.
type Service struct {
	Package     string
	Biz         string
	Service     string
	Methods     []*Method
	GoogleEmpty bool

	UseIO      bool
	UseContext bool
}

// Method is a proto method.
type Method struct {
	Service string
	Name    string
	Request string
	Reply   string
	Uri     string

	// type: unary or stream
	Type MethodType
}

func (s *Service) execute() ([]byte, error) {
	const empty = "google.protobuf.Empty"
	buf := new(bytes.Buffer)
	for _, method := range s.Methods {
		if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
			(method.Type == returnsStreamsType && method.Request == empty) {
			s.GoogleEmpty = true
		}
		if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
			s.UseIO = true
		}
		if method.Type == unaryType {
			s.UseContext = true
		}
	}
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
