package repository

import (
	"fmt"
	"gorm.io/gorm"
	"nistagram/profile/model"
)

type ProfileRepository struct {
	Database *gorm.DB
}

func (repo *ProfileRepository) CreateProfile(profile *model.Profile) error{
	result := repo.Database.Create(profile)
	fmt.Println(result.RowsAffected)
	if result.RowsAffected == 0 {
		return fmt.Errorf("Profile not created")
	}
	fmt.Println("Profile Created")
	return nil
}

func (repo *ProfileRepository) FindInterestByName(name string) model.Interest{
	interest := &model.Interest{}
	repo.Database.Find(&interest, "name", name)
	return *interest
}
