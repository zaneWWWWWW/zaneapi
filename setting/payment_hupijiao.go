package setting

// Hupijiao payment configuration. Runtime availability also requires valid
// credentials, a public HTTPS callback address, and payment compliance.
var (
	HupijiaoEnabled     bool
	HupijiaoEndpoint    = "https://api.dpweixin.com/payment/do.html"
	HupijiaoDisplayName = "支付宝"
	HupijiaoIcon        = "SiAlipay"
	HupijiaoAppID       string
	HupijiaoAppSecret   string
)
