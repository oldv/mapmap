package asm

import (
	"github.com/oldv/mapmap/demo/domain"
	"github.com/oldv/mapmap/demo/dto"
)

// mapmap:assembler
type UserAssembler interface {
	// mapmap:source:Name,target:Age
	// mapmap:source:Age,target:Name
	ToAddDTO(user domain.User) (dto.UserAddDTO, error)

	ToAddUser(addDTO dto.UserAddDTO) (domain.User, error)
}
