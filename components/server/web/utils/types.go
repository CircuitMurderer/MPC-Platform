package utils

type VerifyParams struct {
	ID      string `form:"id"`
	Port    string `form:"port"`
	Address string `form:"address"`
	Workers int    `form:"workers"`
	Operate int    `form:"operate"`
	Scale   int    `form:"scale"`
}
