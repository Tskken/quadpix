package quadpix

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/faiface/pixel"
)

// Entities type is a list of Entity
type Entities []*Entity

// Remove the given entity from the list.
//
// If the entity is not found this function will return nil and an error.
func (e Entities) Remove(entity *Entity) (Entities, error) {
	// search list of entity
	for i := range e {
		// check if the given entity is this entity
		if e[i].IsEqual(entity) {
			switch {
			case len(e) == 1: // Len is 1, clear list case
				return e[:0], nil
			case len(e) == i+1: // Item is last item in list case
				return e[:i], nil
			default: // Default case
				return append(e[:i], e[i+1:]...), nil
			}
		}
	}

	return nil, ErrNoEntityFound
}

// Merge takes the given entities and adds them to this entities list and then returns that new list
// with the added non duplicate entities.
func (e Entities) Merge(entities Entities) Entities {
	for i := range entities {
		// Check for duplicates
		if !e.Contains(entities[i]) {
			// Add to new list
			e = append(e, entities[i])
		}
	}
	return e
}

// Contains checks if the given entity is within the list of entities.
func (e Entities) Contains(entity *Entity) bool {
	// Check entities list for entity
	for i := range e {
		// check if given Entity equals this entity
		if e[i].IsEqual(entity) {
			return true
		}
	}
	return false
}

// Intersect checks if any entity within entities intersects with the given pixel.Rect.
func (e Entities) Intersect(rect pixel.Rect) bool {
	// check if any entities returned intersect the given point
	for i := range e {
		// check for intersect
		if e[i].Intersects(rect) {
			return true
		}
	}
	return false
}

// Intersects returns a list of all entities that the given Rect intersects with within the entities list.
func (e Entities) Intersects(rect pixel.Rect) (entities Entities) {
	// check if any entities returned intersect the given point and if they do add them to the return list
	for i := range e {
		// add to list if they intersect
		if e[i].Intersects(rect) {
			entities = append(entities, e[i])
		}
	}
	return
}

// Action is a function type which can be stored in an Entity for future use.
type Action func()

// Entity is the core data stored with in the Quadpix tree.
//
// Entity holds a pixel.Rect as its bounding box and an ID and a list of posable Action functions.
// ID is set to be a random uint64 number on creation of an Entity.
type Entity struct {
	pixel.Rect

	ID      uint64
	Actions []Action
}

// E creates a new Entity with the given pixel.Rect bounding box and a posable list of Action functions.
//
// If you do not want to store any Action functions you can omit it in the function call as it is a variadic argument.
func E(rect pixel.Rect, actions ...Action) *Entity {
	return &Entity{
		ID:      rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
		Rect:    rect,
		Actions: actions,
	}
}

// IsEqual checks if the given entity is equal to this entity.
//
// IsEqual checks two things:
//		- The entities ID
//		- The entity's pixel.Rect bounds with reflect.DeepEqual()
//
// For this function to return true the given entity has to at least
// have the same ID and bounding box as the entity you are checking.
func (e *Entity) IsEqual(entity *Entity) bool {
	return e.ID == entity.ID && reflect.DeepEqual(e.Rect, entity.Rect)
}

func (e *Entity) String() string {
	return fmt.Sprintf("ID: %v, Bounds: %v\n", e.ID, e.Rect)
}
