## go-kratos scaffold


### usage
```bash
# install
go install https://github.com/exuan/kratos-blades/cmd/kratos-blades
kratos-blades -h
NAME:
   kratos-blades - kratos scaffold

USAGE:
   kratos-blades [global options] command [command options] [arguments...]

COMMANDS:
   proto    
   service  
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

### generate proto

```bash
kratos-blades proto -h
NAME:
   kratos-blades proto - 

USAGE:
   kratos-blades proto [command options] [arguments...]

DESCRIPTION:
   Generate proto file

OPTIONS:
   --api-key value, -k value  http auth key name (default: api_key)
   --auth, -a                 generate api auth security option (default: false)
   --gen-ent, -e              generate ent option (default: false)
   --methods value, -m value  service method name, like "User,Login" (,)separated
   --path value, -p value     generate proto file path (default: ./)
   --service value, -s value  service name, like backend (default: service)
   --version value, -v value  (default: v1)

```

#### generate proto with ent to see [protoc-gen-ent](https://github.com/ent/contrib/tree/master/entproto/cmd/protoc-gen-ent)


#### FAQ:
1. go-swagger error with ` proto not found`, you can copy this `third_party/options/ent/opts.proto` to your `third_party`
2. generate edge with custom schema name, you can use [fix-gen-edge-custom-schema-name](https://github.com/exuan/contrib/tree/fix-gen-edge-custom-schema-name) 

## Acknowledgement
[go-kratos](https://github.com/go-kratos/kratos)

[grpc-ecosystem](https://github.com/grpc-ecosystem)

[entgo](https://github.com/ent/ent)

[ent contrib](https://github.com/ent/contrib)