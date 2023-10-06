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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openclarity/vmclarity/pkg/apiserver/database/odatasql/jsonsql"
)

type SubOption struct {
	Name string `json:"Name"`

	Manufacturer *Manufacturer `json:"Manufacturer"`
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
	Brand         string `json:"Brand,omitempty"`
	NumberOfDisks int    `json:"NumberOfDisks,omitempty"`
}

// Radio StereoType.
type Radio struct {
	ObjectType string `json:"ObjectType"`
	Brand      string `json:"Brand,omitempty"`
	Frequency  string `json:"Frequency,omitempty"`
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

	// Example time field
	BuiltOn *time.Time `json:"BuiltOn"`

	// Always null field
	NullComplexField *Engine `json:"NullComplexField"`
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
			"Id":        {FieldType: StringFieldType},
			"ModelName": {FieldType: StringFieldType},
			"Seats":     {FieldType: NumberFieldType},
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
				FieldType: CollectionFieldType,
				CollectionItemMeta: &FieldMeta{
					FieldType:            RelationshipFieldType,
					RelationshipSchema:   "Manufacturer",
					RelationshipProperty: "Id",
				},
			},
			"BuiltOn": {FieldType: DateTimeFieldType},
			"NullComplexField": {
				FieldType:             ComplexFieldType,
				ComplexFieldSchemas:   []string{"Engine"},
				DiscriminatorProperty: "ObjectType",
			},
		},
	},
	"Manufacturer": {
		Table: "manufacturer_rows",
		Fields: map[string]FieldMeta{
			"Id":   {FieldType: StringFieldType},
			"Name": {FieldType: StringFieldType},
			"Address": {
				FieldType:           ComplexFieldType,
				ComplexFieldSchemas: []string{"Address"},
			},
			"Source": {FieldType: StringFieldType},
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
			"Supercharger": {FieldType: StringFieldType},
			"SubOptions": {
				FieldType: CollectionFieldType,
				CollectionItemMeta: &FieldMeta{
					FieldType:           ComplexFieldType,
					ComplexFieldSchemas: []string{"SubOption"},
				},
			},
			"OtherThings": {
				FieldType:          CollectionFieldType,
				CollectionItemMeta: &FieldMeta{FieldType: StringFieldType},
			},
		},
	},
	"SubOption": {
		Fields: map[string]FieldMeta{
			"Name": {FieldType: StringFieldType},
			"Manufacturer": {
				FieldType:            RelationshipFieldType,
				RelationshipSchema:   "Manufacturer",
				RelationshipProperty: "Id",
			},
		},
	},
	"CDPlayer": {
		Fields: map[string]FieldMeta{
			"ObjectType":    {FieldType: StringFieldType},
			"Brand":         {FieldType: StringFieldType},
			"NumberOfDisks": {FieldType: StringFieldType},
		},
	},
	"Radio": {
		Fields: map[string]FieldMeta{
			"ObjectType": {FieldType: StringFieldType},
			"Brand":      {FieldType: StringFieldType},
			"Frequency":  {FieldType: StringFieldType},
		},
	},
	"Address": {
		Fields: map[string]FieldMeta{
			"City":    {FieldType: StringFieldType},
			"Country": {FieldType: StringFieldType},
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

func addCar(db *gorm.DB, model, manuID string, otherManus []string, seats int, supercharger bool, builtOn *time.Time, optionManu1 string, optionManu2 string, idWithJSONEscapedChars bool) (Car, error) {
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

	radio := &StereoType{}
	err = radio.FromRadio(Radio{
		ObjectType: "Radio",
		Brand:      "Samsung",
		Frequency:  "500mhz",
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

	cdPlayer3 := &StereoType{}
	err = cdPlayer3.FromCDPlayer(CDPlayer{
		ObjectType:    "CDPlayer",
		Brand:         "Unknown",
		NumberOfDisks: 20,
	})
	if err != nil {
		return Car{}, err
	}

	if idWithJSONEscapedChars {
		id = fmt.Sprintf("\"%s\"", id)
	}

	car := Car{
		ID:        id,
		ModelName: model,
		Seats:     seats,
		Engine: &Engine{
			Options: &Options{
				Supercharger: supercharger,
				SubOptions: []SubOption{
					{
						Name: "bluePaint",
						Manufacturer: &Manufacturer{
							ID: optionManu1,
						},
					},
					{
						Name: "blueShoes",
						Manufacturer: &Manufacturer{
							ID: optionManu1,
						},
					},
					{
						Name: "greenPaint",
						Manufacturer: &Manufacturer{
							ID: optionManu2,
						},
					},
					{
						Name: "yellowPaint",
						Manufacturer: &Manufacturer{
							ID: optionManu2,
						},
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
			*cdPlayer3,
		},
		Manufacturer: &Manufacturer{
			ID: manuID,
		},
		Manufacturers: otherMs,
		BuiltOn:       builtOn,
	}
	carBytes, err := json.Marshal(car)
	if err != nil {
		return Car{}, fmt.Errorf("failed to marshal car: %w", err)
	}
	carRow := CarRow{Data: carBytes}
	db.Create(&carRow)
	return car, nil
}

func fromCDPlayer(t *testing.T, cdp CDPlayer) StereoType {
	t.Helper()
	st := &StereoType{}
	err := st.FromCDPlayer(cdp)
	if err != nil {
		t.Fatalf("failed to create stereo: %v", err)
	}
	return *st
}

func fromRadio(t *testing.T, cdp Radio) StereoType {
	t.Helper()
	st := &StereoType{}
	err := st.FromRadio(cdp)
	if err != nil {
		t.Fatalf("failed to create stereo: %v", err)
	}
	return *st
}

// nolint:cyclop,maintidx
func TestBuildSQLQuery(t *testing.T) {
	dbLogger := logger.Default
	dbLogger = dbLogger.LogMode(logger.Info)

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("Failed to create tmp dir for database: %v", err)
	}
	defer os.RemoveAll(dir)

	dbpath := path.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(
		CarRow{},
		ManufacturerRow{},
	); err != nil {
		t.Fatalf("failed to run auto migration: %v", err)
	}

	indexCmd := db.Exec("CREATE INDEX IF NOT EXISTS car_rows_id_idx ON car_rows((Data -> '$.Id'))")
	if indexCmd.Error != nil {
		t.Fatalf("failed to create index cars_id_idx: %v", indexCmd.Error)
	}

	oldtime, err := time.Parse(time.RFC3339, "2021-03-21T08:50:00+00:00")
	if err != nil {
		t.Fatalf("failed to parse old time: %v", err)
	}

	oldtimeAltFormat, err := time.Parse(time.RFC3339, "2021-03-21T08:50:00Z")
	if err != nil {
		t.Fatalf("failed to parse old time alternative format: %v", err)
	}

	oldtimeDiffTz, err := time.Parse(time.RFC3339, "2021-03-21T07:50:00-01:00")
	if err != nil {
		t.Fatalf("failed to parse old time alternative format: %v", err)
	}

	inbetweentime, err := time.Parse(time.RFC3339, "2022-03-21T08:50:00Z")
	if err != nil {
		t.Fatalf("failed to parse in-between time: %v", err)
	}

	newtime, err := time.Parse(time.RFC3339, "2023-03-21T08:50:00+08:00")
	if err != nil {
		t.Fatalf("failed to parse new time: %v", err)
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

	car1, err := addCar(db, "model1", manu1.ID, []string{}, 12, false, &oldtime, manu1.ID, manu2.ID, false)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car2, err := addCar(db, "model2", manu1.ID, []string{manu2.ID, manu3.ID}, 5, true, &oldtime, manu1.ID, manu2.ID, false)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car3, err := addCar(db, "model3", manu2.ID, []string{}, 2, false, &newtime, manu2.ID, manu3.ID, false)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	car4, err := addCar(db, "model4", manu3.ID, []string{}, 2, true, nil, manu2.ID, manu3.ID, true)
	if err != nil {
		t.Errorf("failed to add car to db %v", err)
	}

	type args struct {
		filterString  *string
		selectString  *string
		expandString  *string
		orderbyString *string
		top           *int
		skip          *int
	}
	tests := []struct {
		name    string
		args    args
		want    []Car
		wantErr bool
	}{
		{
			name: "All cars, no filter, select, expand or pagination",
			want: []Car{car1, car2, car3, car4},
		},
		{
			name: "eq filter by primitive type ModelName",
			args: args{
				filterString: PointerTo("ModelName eq 'model1'"),
			},
			want: []Car{car1},
		},
		{
			name: "gt by primitive type Seats",
			args: args{
				filterString: PointerTo("Seats gt 2"),
			},
			want: []Car{car1, car2},
		},
		{
			name: "gt by primitive type on Seats with float",
			args: args{
				filterString: PointerTo("Seats gt 2.0"),
			},
			want: []Car{car1, car2},
		},
		{
			name: "gt by primitive type Seats with no results",
			args: args{
				filterString: PointerTo("Seats gt 14"),
			},
			want: []Car{},
		},
		{
			name: "gt by primitive type with float on Seats with no results",
			args: args{
				filterString: PointerTo("Seats gt 14.0"),
			},
			want: []Car{},
		},
		{
			name: "combined 'and' filter",
			args: args{
				filterString: PointerTo("Seats gt 2 and ModelName eq 'model2'"),
			},
			want: []Car{car2},
		},
		{
			name: "combined 'and' filter with no results",
			args: args{
				filterString: PointerTo("Seats gt 2 and ModelName eq 'doesnotexist'"),
			},
			want: []Car{},
		},
		{
			name: "combined 'or' filter",
			args: args{
				filterString: PointerTo("ModelName eq 'model3' or Seats eq 5"),
			},
			want: []Car{car2, car3},
		},
		{
			name: "'contains' filter",
			args: args{
				filterString: PointerTo("contains(ModelName, '1')"),
			},
			want: []Car{car1},
		},
		{
			name: "'startswith' filter",
			args: args{
				filterString: PointerTo(fmt.Sprintf("startswith(Manufacturer/Id, '%s')", manu1.ID[0:3])),
			},
			want: []Car{car1, car2},
		},
		{
			name: "'endswith' filter",
			args: args{
				filterString: PointerTo("endswith(ModelName, '3')"),
			},
			want: []Car{car3},
		},
		{
			name: "filter on nested field",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Engine/Manufacturer/Id eq '%s'", manu1.ID)),
			},
			want: []Car{car1, car2},
		},
		{
			name: "mismatched brackets",
			args: args{
				filterString: PointerTo("contains(ModelName, '1'"),
			},
			wantErr: true,
		},
		{
			name: "non-boolean filter",
			args: args{
				filterString: PointerTo("ModelName eq"),
			},
			wantErr: true,
		},
		{
			name: "simple select one field",
			args: args{
				selectString: PointerTo("ModelName"),
			},
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
			name: "simple select + select nested field",
			args: args{
				selectString: PointerTo("Seats,Engine/Manufacturer"),
			},
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
			name: "select with nested filter",
			args: args{
				selectString: PointerTo("ModelName,Engine/Options/SubOptions($select=Name;$filter=contains(Name, 'blue'))"),
			},
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
			name: "select fields from nested list",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("OtherStereos/Brand"),
			},
			want: []Car{
				{
					OtherStereos: []StereoType{
						fromRadio(t, Radio{Brand: "Samsung"}),
						fromCDPlayer(t, CDPlayer{Brand: "Unknown"}),
						fromCDPlayer(t, CDPlayer{Brand: "Unknown"}),
					},
				},
			},
		},
		{
			name: "select fields from nested list",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("OtherStereos($select=Brand)"),
			},
			want: []Car{
				{
					OtherStereos: []StereoType{
						fromRadio(t, Radio{Brand: "Samsung"}),
						fromCDPlayer(t, CDPlayer{Brand: "Unknown"}),
						fromCDPlayer(t, CDPlayer{Brand: "Unknown"}),
					},
				},
			},
		},
		{
			name: "select fields from discriminator type",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("MainStereo/NumberOfDisks"),
			},
			want: []Car{
				{
					MainStereo: fromCDPlayer(t, CDPlayer{NumberOfDisks: 12}),
				},
			},
		},
		{
			name: "filter on id with select on id",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("Id"),
			},
			want: []Car{
				{
					ID: car1.ID,
				},
			},
		},
		{
			name: "expand on relationship with select",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("Id,ModelName,Manufacturer"),
				expandString: PointerTo("Manufacturer"),
			},
			want: []Car{
				{
					ID:           car1.ID,
					ModelName:    car1.ModelName,
					Manufacturer: &manu1,
				},
			},
		},
		{
			name: "expand on relationship, relationship not in select",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("Id,ModelName"),
				expandString: PointerTo("Manufacturer"),
			},
			want: []Car{
				{
					ID:           car1.ID,
					ModelName:    car1.ModelName,
					Manufacturer: &manu1,
				},
			},
		},
		{
			name: "expand on relationship with nested select",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("Id,ModelName"),
				expandString: PointerTo("Manufacturer($select=Name)"),
			},
			want: []Car{
				{
					ID:        car1.ID,
					ModelName: car1.ModelName,
					Manufacturer: &Manufacturer{
						Name: manu1.Name,
					},
				},
			},
		},
		{
			name: "expand on relationship collection",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
				selectString: PointerTo("Id,ModelName"),
				expandString: PointerTo("Manufacturers"),
			},
			want: []Car{
				{
					ID:        car2.ID,
					ModelName: car2.ModelName,
					Manufacturers: []Manufacturer{
						manu2,
						manu3,
					},
				},
			},
		},
		{
			name: "expand on relationship collection with filter",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
				selectString: PointerTo("Id,ModelName"),
				expandString: PointerTo("Manufacturers($filter=Name eq 'manu2')"),
			},
			want: []Car{
				{
					ID:        car2.ID,
					ModelName: car2.ModelName,
					Manufacturers: []Manufacturer{
						manu2,
					},
				},
			},
		},
		{
			name: "expand on relationship within complex property",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
				selectString: PointerTo("ModelName,Engine/Manufacturer"),
				expandString: PointerTo("Engine/Manufacturer"),
			},
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
			name: "expand mismatched brackets",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
				selectString: PointerTo("ModelName,Engine/Manufacturer"),
				expandString: PointerTo("Engine/Manufacturer($filter=Name eq 'foo'"),
			},
			want:    []Car{},
			wantErr: true,
		},
		{
			name: "expand bad nested filter",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car2.ID)),
				selectString: PointerTo("ModelName,Engine/Manufacturer"),
				expandString: PointerTo("Engine/Manufacturer($filter=Name eq)"),
			},
			want:    []Car{},
			wantErr: true,
		},
		{
			name: "orderby number of seats decending",
			args: args{
				orderbyString: PointerTo("Seats desc"),
			},
			want: []Car{car1, car2, car3, car4},
		},
		{
			name: "orderby number of seats ascending",
			args: args{
				orderbyString: PointerTo("Seats asc"),
			},
			want: []Car{car3, car4, car2, car1},
		},
		{
			name: "orderby number of seats ascending (direction not specified by user)",
			args: args{
				orderbyString: PointerTo("Seats"),
			},
			want: []Car{car3, car4, car2, car1},
		},
		{
			name: "orderby model name decending",
			args: args{
				orderbyString: PointerTo("ModelName desc"),
			},
			want: []Car{car4, car3, car2, car1},
		},
		{
			name: "order selected list ascending",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("OtherStereos($orderby=NumberOfDisks asc)"),
			},
			want: []Car{
				{
					OtherStereos: []StereoType{
						car1.OtherStereos[0], // radio number of disks is null
						car1.OtherStereos[2], // cdplayer3 number of disks is 20
						car1.OtherStereos[1], // cdplayer2 number of disks is 50
					},
				},
			},
		},
		{
			name: "order selected list decending",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("OtherStereos($orderby=NumberOfDisks desc)"),
			},
			want: []Car{
				{
					OtherStereos: []StereoType{
						car1.OtherStereos[1], // cdplayer2 number of disks is 50
						car1.OtherStereos[2], // cdplayer3 number of disks is 20
						car1.OtherStereos[0], // radio number of disks is null
					},
				},
			},
		},
		{
			name: "orderby sub-object field ascending",
			args: args{
				orderbyString: PointerTo("Engine/Options/Supercharger asc"),
			},
			want: []Car{car1, car3, car2, car4},
		},
		{
			name: "orderby two fields",
			args: args{
				orderbyString: PointerTo("Engine/Options/Supercharger asc, ModelName desc"),
			},
			want: []Car{car3, car1, car4, car2},
		},
		{
			name: "select collection of primitives",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
				selectString: PointerTo("ModelName,Engine/Options/OtherThings"),
			},
			want: []Car{
				{
					ModelName: car1.ModelName,
					Engine: &Engine{
						Options: &Options{
							OtherThings: []string{
								"thing1",
								"thing2",
							},
						},
					},
				},
			},
		},
		{
			name: "filter by time less than",
			args: args{
				filterString: PointerTo(fmt.Sprintf("BuiltOn lt %v", inbetweentime.Format(time.RFC3339))),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter by time greater than",
			args: args{
				filterString: PointerTo(fmt.Sprintf("BuiltOn gt %v", inbetweentime.Format(time.RFC3339))),
			},
			want: []Car{
				car3,
			},
		},
		{
			name: "filter by time equal",
			args: args{
				filterString: PointerTo(fmt.Sprintf("BuiltOn eq %v", oldtime.Format(time.RFC3339))),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter by time equal alternative format",
			args: args{
				filterString: PointerTo(fmt.Sprintf("BuiltOn eq %v", oldtimeAltFormat.Format(time.RFC3339))),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter by time equal diff tz",
			args: args{
				filterString: PointerTo(fmt.Sprintf("BuiltOn eq %v", oldtimeDiffTz.Format(time.RFC3339))),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter on expanded property",
			args: args{
				selectString: PointerTo("ModelName,Manufacturer/Name"),
				filterString: PointerTo("Manufacturer/Name eq 'manu2'"),
				expandString: PointerTo("Manufacturer"),
			},
			want: []Car{
				{
					ModelName: car3.ModelName,
					Manufacturer: &Manufacturer{
						Name: manu2.Name,
					},
				},
			},
		},
		{
			name: "filter on expanded property not selected",
			args: args{
				selectString: PointerTo("ModelName,Manufacturer/Id"),
				filterString: PointerTo("Manufacturer/Name eq 'manu2'"),
				expandString: PointerTo("Manufacturer"),
			},
			want: []Car{
				{
					ModelName: car3.ModelName,
					Manufacturer: &Manufacturer{
						ID: manu2.ID,
					},
				},
			},
		},
		{
			name: "filter on expanded property not expanded or selected",
			args: args{
				selectString: PointerTo("ModelName"),
				filterString: PointerTo("Manufacturer/Name eq 'manu2'"),
			},
			want: []Car{
				{
					ModelName: car3.ModelName,
				},
			},
		},
		{
			name: "filter selected collection on expand entity",
			args: args{
				selectString: PointerTo("ModelName,Engine/Options/SubOptions($select=Name;$filter=Manufacturer/Name eq 'manu1')"),
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car1.ID)),
			},
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
			},
		},
		{
			name: "order by expanded property",
			args: args{
				orderbyString: PointerTo("Manufacturer/Name desc"),
			},
			want: []Car{
				car4,
				car3,
				car1,
				car2,
			},
		},
		{
			name: "get object where field includes json escaped chars",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Id eq '%s'", car4.ID)),
			},
			want: []Car{
				car4,
			},
		},
		{
			name: "filter if list contains an item",
			args: args{
				filterString: PointerTo("Engine/Options/SubOptions/any(o:o/Name eq 'bluePaint')"),
			},
			want: []Car{
				car1,
				car2,
				car3,
				car4,
			},
		},
		{
			name: "filter if list contains an item, no match",
			args: args{
				filterString: PointerTo("Engine/Options/SubOptions/any(o:o/Name eq 'yellow')"),
			},
			want: []Car{},
		},
		{
			name: "filter if list contains an item with expanded properties",
			args: args{
				filterString: PointerTo("Engine/Options/SubOptions/any(o:o/Manufacturer/Name eq 'manu1')"),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter if list contains an item, multi-part query",
			args: args{
				filterString: PointerTo("Engine/Options/SubOptions/any(o:o/Manufacturer/Name eq 'manu1' and o/Name eq 'bluePaint')"),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "filter if list contains an item as part of a larger query",
			args: args{
				filterString: PointerTo("ModelName eq 'model1' and Engine/Options/SubOptions/any(o:o/Manufacturer/Name eq 'manu1')"),
			},
			want: []Car{
				car1,
			},
		},
		{
			name: "negated filter if list contains an item with expanded properties",
			args: args{
				filterString: PointerTo("not Engine/Options/SubOptions/any(o:o/Manufacturer/Name eq 'manu1')"),
			},
			want: []Car{
				car3,
				car4,
			},
		},
		{
			name: "negated eq filter by primitive type ModelName",
			args: args{
				filterString: PointerTo("not (ModelName eq 'model1')"),
			},
			want: []Car{
				car2,
				car3,
				car4,
			},
		},
		{
			name: "if every item in list matches",
			args: args{
				filterString: PointerTo(fmt.Sprintf("Engine/Options/SubOptions/all(o: o/Manufacturer/Id eq '%s' or o/Manufacturer/Id eq '%s')", manu1.ID, manu2.ID)),
			},
			want: []Car{
				car1,
				car2,
			},
		},
		{
			name: "if list length matches",
			args: args{
				filterString: PointerTo("length(Engine/Options/SubOptions) eq 4"),
			},
			want: []Car{
				car1,
				car2,
				car3,
				car4,
			},
		},
		{
			name: "if list length doesn't match",
			args: args{
				filterString: PointerTo("length(Engine/Options/SubOptions) eq 2"),
			},
			want: []Car{},
		},
		{
			name: "eq filter by primitive type ModelName inputs reversed",
			args: args{
				filterString: PointerTo("'model1' eq ModelName"),
			},
			want: []Car{car1},
		},
		{
			name: "eq filter comparing two fields",
			args: args{
				filterString: PointerTo("Manufacturer/Name eq Engine/Manufacturer/Name"),
			},
			want: []Car{
				car1,
				car2,
				car3,
				car4,
			},
		},
		{
			name: "if list ne null",
			args: args{
				filterString: PointerTo("Engine/Options/SubOptions ne null"),
			},
			want: []Car{
				car1,
				car2,
				car3,
				car4,
			},
		},
		{
			name: "if complex type eq null",
			args: args{
				filterString: PointerTo("NullComplexField eq null"),
			},
			want: []Car{
				car1,
				car2,
				car3,
				car4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := BuildSQLQuery(jsonsql.SQLite, carSchemaMetas, "Car", tt.args.filterString, tt.args.selectString, tt.args.expandString, tt.args.orderbyString, tt.args.top, tt.args.skip)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildSQLQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			carResult, err := findCars(db, query)
			if err != nil {
				t.Errorf("BuildSQLQuery() failed to fetch cars: %v", err)
				return
			}

			if diff := cmp.Diff(tt.want, carResult, cmp.AllowUnexported(StereoType{})); diff != "" {
				t.Errorf("BuildSQLQuery() Car mismatch (-want +got):\n%s", diff)
			}
		})
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
