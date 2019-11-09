package auth

type CryptoStringGenerator struct {
}

func NewCryptoStringGenerator() *CryptoStringGenerator {
	return &CryptoStringGenerator{}
}

func (g *CryptoStringGenerator) Generate(bytes int) (string, error) {
	return generateCryptoString(bytes)
}
