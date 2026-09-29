// internal/usecase/generate.go
package usecase

// ports.go のインターフェースから、gomock 用のモックを生成する（go generate ./... で実行する）
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=ports.go -destination=mock/ports_mock.go -package=mock
