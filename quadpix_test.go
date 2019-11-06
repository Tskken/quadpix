package quadpix

import (
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestNew(t *testing.T) {
	type args struct {
		width, height float64
		maxEntities   uint64
		maxDepth      uint16
	}
	tests := []struct {
		name string
		args args
		want *Quadpix
	}{
		{
			name: "create new Quadpix",
			args: args{
				width:       800,
				height:      600,
				maxEntities: 10,
				maxDepth:    4,
			},
			want: &Quadpix{
				node: &node{
					rect:     pixel.R(0, 0, 800, 600),
					entities: make(Entities, 0, 10),
					children: make([]*node, 0, 4),
					depth:    0,
				},
				maxDepth: 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.width, tt.args.height, tt.args.maxEntities, tt.args.maxDepth)
			if got.maxDepth != tt.want.maxDepth {
				t.Errorf("quadpix.New() for maxDepth = %v, want %v", got.maxDepth, tt.want.maxDepth)
			} else if cap(got.entities) != cap(tt.want.entities) {
				t.Errorf("quadpixNew() for maxEntities = %v, want %v", cap(got.entities), cap(tt.want.entities))
			} else if got.rect != tt.want.rect {
				t.Errorf("quadpix.New() for bounds = %v, want %v", got.rect, tt.want.rect)
			}
		})
	}
}

func TestQuadGo_Insert(t *testing.T) {
	type fields struct {
		quadpix *Quadpix
	}
	type args struct {
		rect pixel.Rect
	}
	tests := []struct {
		name   string
		fields fields
		args   []args
		want   Entities
	}{
		{
			name: "basic insert on empty list",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
			},
			args: []args{
				{
					pixel.R(0, 0, 50, 50),
				},
			},
			want: Entities{
				E(pixel.R(0, 0, 50, 50)),
			},
		},
		{
			name: "insert with a split",
			fields: fields{
				quadpix: New(800, 600, 2, 4),
			},
			args: []args{
				{
					pixel.R(0, 0, 50, 50),
				},
				{
					pixel.R(20, 20, 40, 40),
				},
				{
					pixel.R(25, 25, 70, 70),
				},
			},
			want: Entities{
				E(pixel.R(0, 0, 50, 50)),
				E(pixel.R(20, 20, 40, 40)),
				E(pixel.R(25, 25, 70, 70)),
			},
		},
		{
			name: "insert with no split for max depth",
			fields: fields{
				quadpix: New(800, 600, 2, 0),
			},
			args: []args{
				{
					pixel.R(0, 0, 50, 50),
				},
				{
					pixel.R(20, 20, 40, 40),
				},
				{
					pixel.R(25, 25, 70, 70),
				},
			},
			want: Entities{
				E(pixel.R(0, 0, 50, 50)),
				E(pixel.R(20, 20, 40, 40)),
				E(pixel.R(25, 25, 70, 70)),
			},
		},
		{
			name: "inert 4 quadrents",
			fields: fields{
				quadpix: New(800, 600, 1, 4),
			},
			args: []args{
				{
					pixel.R(0, 0, 50, 50),
				},
				{
					pixel.R(0, 350, 50, 500),
				},
				{
					pixel.R(450, 0, 600, 50),
				},
				{
					pixel.R(450, 350, 600, 500),
				},
			},
			want: Entities{
				E(pixel.R(0, 0, 50, 50)),
				E(pixel.R(0, 350, 50, 500)),
				E(pixel.R(450, 0, 600, 50)),
				E(pixel.R(450, 350, 600, 500)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				tt.fields.quadpix.Insert(arg.rect)
			}

			for _, wnt := range tt.want {
				found := false

				for _, e := range <-tt.fields.quadpix.Retrieve(wnt.Rect) {
					if e.Rect == wnt.Rect {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("QuadGo.Insert() could not find %v in tree", wnt)
				}
			}
		})
	}
}

func TestQuadGo_InsertEntities(t *testing.T) {
	type fields struct {
		quadpix *Quadpix
	}
	type args struct {
		entities Entities
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Entities
		wantErr error
	}{
		{
			name: "insert 1 entity",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
			},
			args: args{
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
				},
			},
			want: Entities{
				&Entity{
					ID:      1,
					Rect:    pixel.R(0, 0, 50, 50),
					Actions: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "insert no entities error",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
			},
			args: args{
				entities: nil,
			},
			want:    nil,
			wantErr: ErrNoEntitiesGiven,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.args.entities...)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("QuadGo.InsertEntities() unwanted error type = %v, want %v", err, tt.wantErr)
			}

			for _, ent := range tt.want {
				if !<-tt.fields.quadpix.IsEntity(ent) {
					t.Errorf("QuadGo.InsertEntities() entity not inserted %v", ent)
				}
			}
		})
	}
}

func TestQuadGo_Remove(t *testing.T) {
	type fields struct {
		quadpix  *Quadpix
		entities Entities
	}
	type args struct {
		entity *Entity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove 1 entity",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      2,
						Rect:    pixel.R(20, 20, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      3,
						Rect:    pixel.R(5, 5, 90, 80),
						Actions: nil,
					},
				},
			},
			args: args{
				&Entity{
					ID:      2,
					Rect:    pixel.R(20, 20, 50, 50),
					Actions: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove and collapse",
			fields: fields{
				quadpix: New(800, 600, 2, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      2,
						Rect:    pixel.R(25, 25, 50, 60),
						Actions: nil,
					},
					&Entity{
						ID:      3,
						Rect:    pixel.R(5, 5, 90, 80),
						Actions: nil,
					},
				},
			},
			args: args{
				&Entity{
					ID:      1,
					Rect:    pixel.R(0, 0, 50, 50),
					Actions: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove non entity error",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					E(pixel.R(20, 20, 50, 50)),
					E(pixel.R(5, 5, 90, 80)),
				},
			},
			args: args{
				E(pixel.R(0, 0, 50, 50)),
			},
			wantErr: ErrNoEntityFound,
		},
		{
			name: "remove non entity error from leafs",
			fields: fields{
				quadpix: New(800, 600, 1, 4),
				entities: Entities{
					E(pixel.R(20, 20, 50, 50)),
					E(pixel.R(5, 5, 90, 80)),
				},
			},
			args: args{
				E(pixel.R(0, 0, 50, 50)),
			},
			wantErr: ErrNoEntityFound,
		},
		{
			name: "remove last entity from list",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
				},
			},
			args: args{
				&Entity{
					ID:      1,
					Rect:    pixel.R(0, 0, 50, 50),
					Actions: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove last entity from end of list",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      2,
						Rect:    pixel.R(20, 20, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      3,
						Rect:    pixel.R(5, 5, 90, 80),
						Actions: nil,
					},
				},
			},
			args: args{
				&Entity{
					ID:      3,
					Rect:    pixel.R(5, 5, 90, 80),
					Actions: nil,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.Remove() insert entities with error %v", err)
			}

			err = tt.fields.quadpix.Remove(tt.args.entity)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("QuadGo.Remove() got an unwanted error = %v, want %v", err, tt.wantErr)
			}

			if <-tt.fields.quadpix.IsEntity(tt.args.entity) {
				t.Errorf("QuadGo.Remove() found entity even after delete")
			}
		})
	}
}

func TestQuadGo_Retrieve(t *testing.T) {
	type fields struct {
		quadpix  *Quadpix
		entities Entities
	}
	type args struct {
		rect pixel.Rect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "find 1 value",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:   1,
						Rect: pixel.R(0, 0, 50, 50),
					},
				},
			},
			args: args{
				rect: pixel.R(5, 5, 10, 10),
			},
			want: Entities{
				&Entity{
					ID:   1,
					Rect: pixel.R(0, 0, 50, 50),
				},
			},
		},
		{
			name: "find 1 value from child",
			fields: fields{
				quadpix: New(800, 600, 2, 4),
				entities: Entities{
					&Entity{
						ID:   1,
						Rect: pixel.R(0, 0, 50, 50),
					},
					&Entity{
						ID:   2,
						Rect: pixel.R(500, 400, 700, 600),
					},
					&Entity{
						ID:   3,
						Rect: pixel.R(450, 350, 600, 550),
					},
				},
			},
			args: args{
				rect: pixel.R(5, 5, 10, 10),
			},
			want: Entities{
				&Entity{
					ID:   1,
					Rect: pixel.R(0, 0, 50, 50),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.Retrieve() got error on insert %v", err)
			}

			entities := <-tt.fields.quadpix.Retrieve(tt.args.rect)

			for _, ent := range tt.want {
				if !entities.Contains(ent) {
					t.Errorf("QuadGo.Retrieve() wanted value not found, entities: %v, want: %v", entities, ent)
				}
			}
		})
	}
}

func TestQuadGo_IsEntity(t *testing.T) {
	type fields struct {
		quadpix  *Quadpix
		entities Entities
	}
	type args struct {
		entity *Entity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is entity true",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:   1,
						Rect: pixel.R(0, 0, 50, 50),
					},
				},
			},
			args: args{
				entity: &Entity{
					ID:   1,
					Rect: pixel.R(0, 0, 50, 50),
				},
			},
			want: true,
		},
		{
			name: "is entity false",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
				},
			},
			args: args{
				entity: E(pixel.R(10, 10, 50, 50)),
			},
			want: false,
		},
		{
			name: "is entity true from branch",
			fields: fields{
				quadpix: New(800, 800, 2, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      2,
						Rect:    pixel.R(25, 25, 50, 60),
						Actions: nil,
					},
					&Entity{
						ID:      3,
						Rect:    pixel.R(5, 5, 90, 80),
						Actions: nil,
					},
				},
			},
			args: args{
				entity: &Entity{
					ID:      1,
					Rect:    pixel.R(0, 0, 50, 50),
					Actions: nil,
				},
			},
			want: true,
		},
		{
			name: "is entity false from branch",
			fields: fields{
				quadpix: New(800, 800, 2, 4),
				entities: Entities{
					&Entity{
						ID:      1,
						Rect:    pixel.R(0, 0, 50, 50),
						Actions: nil,
					},
					&Entity{
						ID:      2,
						Rect:    pixel.R(25, 25, 50, 60),
						Actions: nil,
					},
					&Entity{
						ID:      3,
						Rect:    pixel.R(5, 5, 90, 80),
						Actions: nil,
					},
				},
			},
			args: args{
				entity: &Entity{
					ID:      5,
					Rect:    pixel.R(5, 5, 50, 50),
					Actions: nil,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsEntity() got error on insert %v", err)
			}

			if got := <-tt.fields.quadpix.IsEntity(tt.args.entity); got != tt.want {
				t.Errorf("QuadGo.IsEntity() = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestQuadGo_IsIntersect(t *testing.T) {
	type fields struct {
		quadpix  *Quadpix
		entities Entities
	}
	type args struct {
		rect pixel.Rect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
				},
			},
			args: args{
				rect: pixel.R(5, 5, 10, 10),
			},
			want: true,
		},
		{
			name: "is intersect false",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
				},
			},
			args: args{
				rect: pixel.R(60, 60, 70, 70),
			},
			want: false,
		},
		{
			name: "is intersect through leafs",
			fields: fields{
				quadpix: New(800, 600, 1, 2),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
					E(pixel.R(100, 100, 400, 400)),
				},
			},
			args: args{
				rect: pixel.R(50, 50, 70, 70),
			},
			want: true,
		},
		{
			name: "is intersect through leafs",
			fields: fields{
				quadpix: New(800, 600, 1, 2),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
					E(pixel.R(100, 100, 400, 400)),
				},
			},
			args: args{
				rect: pixel.R(60, 60, 70, 70),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersect() got error on insert %v", err)
			}

			if got := <-tt.fields.quadpix.Intersect(tt.args.rect); got != tt.want {
				t.Errorf("QuadGo.IsIntersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuadGo_Intersects(t *testing.T) {
	type fields struct {
		quadpix  *Quadpix
		entities Entities
	}
	type args struct {
		rect pixel.Rect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					&Entity{
						ID:   1,
						Rect: pixel.R(0, 0, 50, 50),
					},
				},
			},
			args: args{
				rect: pixel.R(5, 5, 10, 10),
			},
			want: Entities{
				&Entity{
					ID:   1,
					Rect: pixel.R(0, 0, 50, 50),
				},
			},
		},
		{
			name: "is intersect false",
			fields: fields{
				quadpix: New(800, 600, 10, 4),
				entities: Entities{
					E(pixel.R(0, 0, 50, 50)),
				},
			},
			args: args{
				rect: pixel.R(60, 60, 70, 70),
			},
			want: Entities{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadpix.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersect() got error on insert %v", err)
			}

			got := <-tt.fields.quadpix.Intersects(tt.args.rect)

			if len(tt.want) == 0 && len(got) != 0 {
				t.Errorf("QuadGo.Intersects() wanted no intersects but got %v", got)
			} else {
				for _, ent := range tt.want {
					if !got.Contains(ent) {
						t.Errorf("QuadGo.Intersects() did not return wanted entity %v", ent)
					}
				}
			}
		})
	}
}
