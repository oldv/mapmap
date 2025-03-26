package asm

import (
	"mapmap/demo/domain"
	"mapmap/demo/dto"
)

// mapmap:assembler
type UserAssembler interface {
	// mapmap:source:Name,target:Age
	// mapmap:source:Age,target:Name
	ToAddDTO(user domain.User) (dto.UserAddDTO, error)

	ToAddUser(addDTO dto.UserAddDTO) (domain.User, error)
}
