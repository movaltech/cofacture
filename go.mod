module github.com/movaltech/cofacture/v2

go 1.26.5

require (
	github.com/beevik/etree v1.6.0
	github.com/google/uuid v1.6.0
	github.com/movaltech/decimal v1.0.1
	github.com/russellhaering/goxmldsig v1.6.0
	golang.org/x/crypto v0.53.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
)

retract (
	// Tagged on the 2.8.1 commit by mistake and lacks its changes; use v2.10.0.
	v2.9.0
	// Tagged on the 2.1.0 commit by mistake and lacks its changes; use v2.2.1.
	v2.2.0
)
