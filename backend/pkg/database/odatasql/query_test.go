// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package odatasql

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SubOption struct {
	Name string `json:"Name"`
}

type Options struct {
	// Example field of a primitve type inside of a nested complex type
	Supercharger bool `json:"Supercharger"`

	// Example field of a collection of complex type inside of a nested
	// complex type
	SubOptions []SubOption `json:"SubOptions"`

	// Example field of a collection of primitive type inside of a nested
	// complex type
	OtherThings []string `json:OtherThings`
}

type Engine struct {
	// Example field of a complex type inside of a complex type
	Options *Options `json:"Options"`

	// Example field which is a relationship to a different collection
	// within a complex property
	Manufacturer *Manufacturer `json:"Manufacturer"`
}

// CDPlayer StereoType.
type CDPlayer struct {
	ObjectType    string `json:"ObjectType"`
	Brand         string `json:"Brand"`
	NumberOfDisks int    `json:"NumberOfDisks"`
}

// Radio StereoType.
type Radio struct {
	ObjectType string `json:"ObjectType"`
	Brand      string `json:"Brand"`
	Frequency  string `json:"Frequency"`
}

type Address struct {
	City    string `json:"City,omitempty"`
	Country string `json:"Country,omitempty"`
}

type StereoType struct {
	union json.RawMessage
}

func (t StereoType) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err // nolint:wrapcheck
}

func (t *StereoType) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err // nolint:wrapcheck
}

func (t *StereoType) FromCDPlayer(v CDPlayer) error {
	v.ObjectType = "CDPlayer"
	b, err := json.Marshal(v)
	t.union = b
	return err // nolint:wrapcheck
}

func (t *StereoType) FromRadio(v Radio) error {
	v.ObjectType = "Radio"
	b, err := json.Marshal(v)
	t.union = b
	return err // nolint:wrapcheck
}

type Manufacturer struct {
	ID      string   `json:"Id"`
	Name    string   `json:"Name,omitempty"`
	Address *Address `json:"Address,omitempty"`
}

type Car struct {
	ID string `json:"Id"`

	// Example field which is a primitive type
	ModelName string `json:"ModelName"`
	Seats     int    `json:"Seats"`

	// Example field which is a complex property
	Engine *Engine `json:"Engine"`

	// Example field which can have different object types
	MainStereo StereoType `json:"MainStereo"`

	// Example field which is a collection of different object types
	OtherStereos []StereoType `json:"OtherStereos"`

	// Example field which is a relationship to a different collection
	Manufacturer *Manufacturer `json:"Manufacturer"`

	// Example field which is a collection of relationships to a
	// different collection
	Manufacturers []Manufacturer `json:"Manufacturers"`
}

// Gorm Table Definitions.
type CarRow struct {
	gorm.Model
	Data datatypes.JSON
}

type ManufacturerRow struct {
	gorm.Model
	Data datatypes.JSON
}

var carSchemaMetas = map[string]SchemaMeta{
	"Car": {
		Table: "car_rows",
		Fields: map[string]FieldMeta{
			"Id":        {FieldType: PrimitiveFieldType},
			"ModelName": {FieldType: PrimitiveFieldType},
			"Seats":     {FieldType: PrimitiveFieldType},
			"Engine": {
				FieldType:           ComplexFieldType,
				ComplexFieldSchemas: []string{"Engine"},
			},
			"MainStereo": {
				FieldType:             ComplexFieldType,
				ComplexFieldSchemas:   []string{"CDPlayer", "Radio"},
				DiscriminatorProperty: "ObjectType",
			},
			"OtherStereos": {
				FieldType: CollectionFieldType,
				CollectionItemMeta: &FieldMeta{
					FieldType:             ComplexFieldType,
					ComplexFieldSchemas:   []string{"CDPlayer", "Radio"},
					DiscriminatorProperty: "ObjectType",
				},
			},
			"Manufacturer": {
				FieldType:            RelationshipFieldType,
				RelationshipSchema:   "Manufacturer",
				RelationshipProperty: "Id",
			},
			"Manufacturers": {
				FieldType:            RelationshipCollectionFieldType,
				RelationshipSchema:   "Manufacturer",
				RelationshipProperty: "Id",
			},
		},
	},
	"Manufacturer": {
		Table: "manufacturer_rows",
		Fields: map[string]FieldMeta{
			"Id":   {FieldType: PrimitiveFieldType},
			"Name": {FieldType: PrimitiveFieldType},
			"Address": {
				FieldType:           ComplexFieldType,
				ComplexFieldSchemas: []string{"Address"},
			},
			"Source": {FieldType: PrimitiveFieldType},
		},
	},
	"Engine": {
		Fields: map[string]FieldMeta{
			"Options": {
				FieldType:           ComplexFieldType,
				ComplexFieldSchemas: []string{"Options"},
			},
			"Manufacturer": {
				FieldType:            RelationshipFieldType,
				RelationshipSchema:   "Manufacturer",
				RelationshipProperty: "Id",
			},
		},
	},
	"Options": {
		Fields: map[string]FieldMeta{
			"Supercharger": {FieldType: PrimitiveFieldType},
			"SubOptions": {
				FieldType: CollectionFieldType,
				CollectionItemMeta: &FieldMeta{
					FieldType:           ComplexFieldType,
					ComplexFieldSchemas: []string{"SubOption"},
				},
			},
			"OtherThings": {
				FieldType:          CollectionFieldType,
				CollectionItemMeta: &FieldMeta{FieldType: PrimitiveFieldType},
			},
		},
	},
	"SubOption": {
		Fields: map[string]FieldMeta{
			"Name": {FieldType: PrimitiveFieldType},
		},
	},
	"CDPlayer": {
		Fields: map[string]FieldMeta{
			"ObjectType":    {FieldType: PrimitiveFieldType},
			"Brand":         {FieldType: PrimitiveFieldType},
			"NumberOfDisks": {FieldType: PrimitiveFieldType},
		},
	},
	"Radio": {
		Fields: map[string]FieldMeta{
			"ObjectType": {FieldType: PrimitiveFieldType},
			"Brand":      {FieldType: PrimitiveFieldType},
			"Frequency":  {FieldType: PrimitiveFieldType},
		},
	},
	"Address": {
		Fields: map[string]FieldMeta{
			"City":    {FieldType: PrimitiveFieldType},
			"Country": {FieldType: PrimitiveFieldType},
		},
	},
}

func addManufacturer(db *gorm.DB, postfix string) (Manufacturer, error) {
	id := uuid.New().String()
	manufacturer := Manufacturer{
		ID:   id,
		Name: fmt.Sprintf("manu%s", postfix),
		Address: &Address{
			City:    fmt.Sprintf("city%s", postfix),
			Country: "middleofnowhere",
		},
	}
	manuBytes, err := json.Marshal(manufacturer)
	if err != nil {
		return Manufacturer{}, fmt.Errorf("failed to marshal manufacturer: %w", err)
	}
	manu := ManufacturerRow{Data: manuBytes}
	db.Create(&manu)
	return manufacturer, nil
}

func addCar(db *gorm.DB, model, manuID string, otherManus []string, seats int) (Car, error) {
	id := uuid.New().String()

	otherMs := []Manufacturer{}
	for _, otherManu := range otherManus {
		otherMs = append(otherMs, Manufacturer{
			ID: otherManu,
		})
	}

	cdPlayer1 := &StereoType{}
	err := cdPlayer1.FromCDPlayer(CDPlayer{
		ObjectType:    "CDPlayer",
		Brand:         "Sony",
		NumberOfDisks: 12,
	})
	if err != nil {
		return Car{}, err
	}

	cdPlayer2 := &StereoType{}
	err = cdPlayer2.FromCDPlayer(CDPlayer{
		ObjectType:    "CDPlayer",
		Brand:         "Unknown",
		NumberOfDisks: 50,
	})
	if err != nil {
		return Car{}, err
	}

	radio := &StereoType{}
	err = radio.FromRadio(Radio{
		ObjectType: "Radio",
		Brand:      "Samsung",
		Frequency:  "500mhz",
	})
	if err != nil {
		return Car{}, err
	}

	car := Car{
		ID:        id,
		ModelName: model,
		Seats:     seats,
		Engine: &Engine{
			Options: &Options{
				Supercharger: false,
				SubOptions: []SubOption{
					{
						Name: "bluePaint",
					},
					{
						Name: "blueShoes",
					},
					{
						Name: "greenPaint",
					},
					{
						Name: "yellowPaint",
					},
				},
				OtherThings: []string{
					"thing1",
					"thing2",
				},
			},
			Manufacturer: &Manufacturer{
				ID: manuID,
			},
		},
		MainStereo: *cdPlayer1,
		OtherStereos: []StereoType{
			*radio,
			*cdPlayer2,
		},
		Manufacturer: &Manufacturer{
			ID: manuID,
		},
		Manufacturers: otherMs,
	}
	carBytes, err := json.Marshal(car)
	if err != nil {
		return Car{}, fmt.Errorf("failed to marshal car: %w", err)
	}
	carRow := CarRow{Data: carBytes}
	db.Create(&carRow)
	return car, nil
}

// nolint:cyclop,maintidx
func Test_BuildSQLQuery(t *testing.T) {
	dbLogger := logger.Default
	dbLogger = dbLogger.LogMode(logger.Info)

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("Failed to create tmp dir for database: %v", err)
	}
	defer os.RemoveAll(dir)

	dbpath := path.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		t.Errorf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(
		CarRow{},
		ManufacturerRow{},
	); err != nil {
		t.Errorf("failed to run auto migration: %v", err)
	}

	manu1, err := addManufacturer(db, "1")
	if err != nil {
		t.Errorf("failed to add manufacturer to db %v", err)
	}

	manu2, err := addManufacturer(db, "2")
	if err != nil {
		t.Errorf("failed to add manufacturer to db %v", err)
	}

	manu3, err := addManufacturer(db, "3")
	if err != nil {
		t.Errorf("failed to add manufacturer to db %v", err)
	}

	car1, err := addCar(db, "model1", manu1.ID, []string{}, 5)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car2, err := addCar(db, "model2", manu1.ID, []string{manu2.ID, manu3.ID}, 5)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car3, err := addCar(db, "model3", manu2.ID, []string{}, 2)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car4, err := addCar(db, "model4", manu3.ID, []string{}, 2)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	tests := []struct {
		name            string
		filterString    *string
		selectString    *string
		expandString    *string
		top             *int
		skip            *int
		want            []Car
		buildQueryError bool
	}{
		{
			name:         "All cars, no filter, select, expand or pagination",
			filterString: nil,
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1, car2, car3, car4},
		},
		{
			name:         "eq filter by primitive type ModelName",
			filterString: PointerTo("ModelName eq 'model1'"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1},
		},
		{
			name:         "gt by primitive type Seats",
			filterString: PointerTo("Seats gt 2"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1, car2},
		},
		{
			name:         "gt by primitive type on Seats with float",
			filterString: PointerTo("Seats gt 2.0"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1, car2},
		},
		{
			name:         "gt by primitive type Seats with no results",
			filterString: PointerTo("Seats gt 10"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{},
		},
		{
			name:         "gt by primitive type with float on Seats with no results",
			filterString: PointerTo("Seats gt 10.0"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{},
		},
		{
			name:         "combined 'and' filter",
			filterString: PointerTo("Seats gt 2 and ModelName eq 'model2'"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car2},
		},
		{
			name:         "combined 'and' filter with no results",
			filterString: PointerTo("Seats gt 2 and ModelName eq 'doesnotexist'"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{},
		},
		{
			name:         "combined 'or' filter",
			filterString: PointerTo("ModelName eq 'model3' or Seats eq 5"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1, car2, car3},
		},
		{
			name:         "'contains' filter",
			filterString: PointerTo("contains(ModelName, '1')"),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1},
		},
		{
			name:         "filter on nested field",
			filterString: PointerTo(fmt.Sprintf("Engine/Manufacturer/Id eq '%s'", manu1.ID)),
			selectString: nil,
			expandString: nil,
			top:          nil,
			skip:         nil,
			want:         []Car{car1, car2},
		},
		{
			name:            "mismatched brackets",
			filterString:    PointerTo("contains(ModelName, '1'"),
			selectString:    nil,
			expandString:    nil,
			top:             nil,
			skip:            nil,
			buildQueryError: true,
		},
		{
			name:            "non-boolean filter",
			filterString:    PointerTo("ModelName eq"),
			selectString:    nil,
			expandString:    nil,
			top:             nil,
			skip:            nil,
			buildQueryError: true,
		},
		{
			name:         "simple select one field",
			filterString: nil,
			selectString: PointerTo("ModelName"),
			expandString: nil,
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car1.ModelName,
				},
				{
					ModelName: car2.ModelName,
				},
				{
					ModelName: car3.ModelName,
				},
				{
					ModelName: car4.ModelName,
				},
			},
		},
		{
			name:         "simple select + select nested field",
			filterString: nil,
			selectString: PointerTo("Seats,Engine/Manufacturer"),
			expandString: nil,
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					Seats: car1.Seats,
					Engine: &Engine{
						Manufacturer: car1.Manufacturer,
					},
				},
				{
					Seats: car2.Seats,
					Engine: &Engine{
						Manufacturer: car2.Manufacturer,
					},
				},
				{
					Seats: car3.Seats,
					Engine: &Engine{
						Manufacturer: car3.Manufacturer,
					},
				},
				{
					Seats: car4.Seats,
					Engine: &Engine{
						Manufacturer: car4.Manufacturer,
					},
				},
			},
		},
		{
			name:         "select with nested filter",
			filterString: nil,
			selectString: PointerTo("ModelName,Engine/Options/SubOptions($filter=contains(Name, 'blue'))"),
			expandString: nil,
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car1.ModelName,
					Engine: &Engine{
						Options: &Options{
							SubOptions: []SubOption{
								{
									Name: "bluePaint",
								},
								{
									Name: "blueShoes",
								},
							},
						},
					},
				},
				{
					ModelName: car2.ModelName,
					Engine: &Engine{
						Options: &Options{
							SubOptions: []SubOption{
								{
									Name: "bluePaint",
								},
								{
									Name: "blueShoes",
								},
							},
						},
					},
				},
				{
					ModelName: car3.ModelName,
					Engine: &Engine{
						Options: &Options{
							SubOptions: []SubOption{
								{
									Name: "bluePaint",
								},
								{
									Name: "blueShoes",
								},
							},
						},
					},
				},
				{
					ModelName: car4.ModelName,
					Engine: &Engine{
						Options: &Options{
							SubOptions: []SubOption{
								{
									Name: "bluePaint",
								},
								{
									Name: "blueShoes",
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "expand on relationship",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
			selectString: PointerTo("ModelName,Manufacturer"),
			expandString: PointerTo("Manufacturer"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName:    car1.ModelName,
					Manufacturer: &manu1,
				},
			},
		},
		{
			name:         "expand on relationship, relationship not in select",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
			selectString: PointerTo("ModelName"),
			expandString: PointerTo("Manufacturer"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName:    car1.ModelName,
					Manufacturer: &manu1,
				},
			},
		},
		{
			name:         "expand on relationship with nested select",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
			selectString: PointerTo("ModelName"),
			expandString: PointerTo("Manufacturer($select=Name)"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car1.ModelName,
					Manufacturer: &Manufacturer{
						Name: manu1.Name,
					},
				},
			},
		},
		{
			name:         "expand on relationship collection",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
			selectString: PointerTo("ModelName"),
			expandString: PointerTo("Manufacturers"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car2.ModelName,
					Manufacturers: []Manufacturer{
						manu2,
						manu3,
					},
				},
			},
		},
		{
			name:         "expand on relationship collection with filter",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
			selectString: PointerTo("ModelName"),
			expandString: PointerTo("Manufacturers($filter=Name eq 'manu2')"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car2.ModelName,
					Manufacturers: []Manufacturer{
						manu2,
					},
				},
			},
		},
		{
			name:         "expand on relationship within complex property",
			filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
			selectString: PointerTo("ModelName,Engine/Manufacturer"),
			expandString: PointerTo("Engine/Manufacturer"),
			top:          nil,
			skip:         nil,
			want: []Car{
				{
					ModelName: car2.ModelName,
					Engine: &Engine{
						Manufacturer: &manu1,
					},
				},
			},
		},
		{
			name:            "expand mismatched brackets",
			filterString:    PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
			selectString:    PointerTo("ModelName,Engine/Manufacturer"),
			expandString:    PointerTo("Engine/Manufacturer($filter=Name eq 'foo'"),
			top:             nil,
			skip:            nil,
			want:            []Car{},
			buildQueryError: true,
		},
		{
			name:            "expand bad nested filter",
			filterString:    PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
			selectString:    PointerTo("ModelName,Engine/Manufacturer"),
			expandString:    PointerTo("Engine/Manufacturer($filter=Name eq)"),
			top:             nil,
			skip:            nil,
			want:            []Car{},
			buildQueryError: true,
		},
	}

	for _, test := range tests {
		query, err := BuildSQLQuery(carSchemaMetas, "Car", test.filterString, test.selectString, test.expandString, test.top, test.skip)
		if err != nil {
			if !test.buildQueryError {
				t.Errorf("[%s] failed to build query for DB: %v", test.name, err)
			}
			continue
		}

		carResult, err := findCars(db, query)
		if err != nil {
			t.Errorf("[%s] failed to fetch cars: %v", test.name, err)
			continue
		}

		if diff := cmp.Diff(test.want, carResult, cmp.AllowUnexported(StereoType{})); diff != "" {
			t.Errorf("[%s] []Car mismatch (-want +got):\n%s", test.name, diff)
		}
	}
}

func findCars(db *gorm.DB, query string) ([]Car, error) {
	var results []CarRow
	if err := db.Raw(query).Find(&results).Error; err != nil {
		return []Car{}, fmt.Errorf("failed to query DB: %w", err)
	}

	carResult := []Car{}
	for _, result := range results {
		var car Car
		err := json.Unmarshal(result.Data, &car)
		if err != nil {
			return []Car{}, fmt.Errorf("failed to unmarshal car data %v: %w", result.Data, err)
		}
		carResult = append(carResult, car)
	}

	return carResult, nil
}

func PointerTo[T any](value T) *T {
	return &value
}
