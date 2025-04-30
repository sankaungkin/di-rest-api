package category

import (
	"reflect"
	"testing"

	"github.com/sankangkin/di-rest-api/internal/models"
)

func TestCategoryRepository_Update(t *testing.T) {
	type args struct {
		input *models.Category
	}
	tests := []struct {
		name    string
		r       *CategoryRepository
		args    args
		want    *models.Category
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Update(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CategoryRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CategoryRepository.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}
