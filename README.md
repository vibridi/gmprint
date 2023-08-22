## gmprint

gmprint (gRPC method print) is a quick-and-dirty command line utility to parse the AST of a valid Go file 
and print gRPC method calls. I use it in some bash scripts to compose a list of interactions between gRPC microservices. 

### Usage

```
gmprint -f <file_name> -c <client_name>
```
where `file_name` is the Go file you want to inspect and `client_name` is the identifier of a struct
that holds a reference to a gRPC client.

### How it works

It parses the AST of the Go file and looks for call expressions with a selector involving `client_name` and prints
them in sorted order. The Go file is assumed to contain a struct that implements methods with gRPC calls.

### Example

Input file (only required content is shown, the file must be valid Go):
```go
package mypackage

import (
	// ...
)

type userService struct {
	client pb.UserServiceClient
}

func (s *userService) GetUser(ctx context.Context, userId uint64) (*pb.Account, error) {
	account, err := s.client.GetUser(
		ctx,
		&pb.UserIdRequest{
			Id: userId,
		},
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *userService) IsBetaUser(ctx context.Context) (bool, error) {
	isBetaUser, err := s.client.IsBetaUser(ctx, &pb.Empty{})
	if err != nil {
		return false, err
	}
	return isBetaUser.Result, nil
}
```
Invoked as:

```
gmprint -f user_service.go -c userService
```

Will print to stdout:

```
GetUser, IsBetaUser
```
which are the methods called on `pb.UserServiceClient`
