// internal/repository/mock_gen.go
package repository

//go:generate mockgen -typed -source=./user.go -destination=../mocks/mock_user_repository.go -package=mocks UserRepositoryIface
//go:generate mockgen -typed -source=./user_factor.go -destination=../mocks/mock_user_factor_repository.go -package=mocks UserFactorRepositoryIface
//go:generate mockgen -typed -source=./organization.go -destination=../mocks/mock_organization_repository.go -package=mocks OrganizationRepositoryIface
