package quadpix

import (
	"github.com/faiface/pixel"
)

// Quadpix is the core structure holding the quadtree data for quadpix.
type Quadpix struct {
	*node

	maxDepth uint16
}

// New creates a new instance of Quadpix with the given arguments.
//
// Args:
//		- Width, height float64: the width and height of the root of the tree.
//		- maxEntities uint64: the max number of entities per node for the tree before the node splits.
//		- maxDepth uin16: the max depth of the tree.
//
// Returns:
//		-Pointer to the newly created Quadpix instance.
func New(width, height float64, maxEntities uint64, maxDepth uint16) *Quadpix {
	return &Quadpix{
		node: &node{
			rect:     pixel.R(0, 0, width, height),
			entities: make(Entities, 0, maxEntities),
			children: make([]*node, 0, 4),
			depth:    0,
		},
		maxDepth: maxDepth,
	}
}

// Insert adds the given pixel.Rect to the tree as an entity bound.
//
// Insert also takes a variadic number of Action functions which can be stored in the Entity.
// These functions could be used for actions done on intersect.
//
// If no Actions are given it will set set to nil.
func (q *Quadpix) Insert(rect pixel.Rect, action ...Action) {
	q.insert(E(rect, action...), q.maxDepth)
}

// InsertEntities inserts any number of Entity's to the tree.
//
// This function will return an error if no entities are given to InsertEntities.
func (q *Quadpix) InsertEntities(entities ...*Entity) error {
	// Check for no entities given.
	if len(entities) == 0 {
		return ErrNoEntitiesGiven
	}

	// Add entities to tree.
	for _, e := range entities {
		q.insert(e, q.maxDepth)
	}

	return nil
}

// Remove the given entity from the tree.
//
// Remove will return an error if the given entity can not be found in the tree.
// Entity's are compared on two values, there UID which is set to a random uint64 number on creation,
// and a reflect.DeepEqual() comparison of the Rect bounds of the entity's.
// If you whish to remove a given entity from the tree you must make sure you have at least the same ID and pixel.Rect
// as the entity you are trying to remove.
func (q *Quadpix) Remove(entity *Entity) error {
	return q.remove(entity)
}

// Retrieve gets all entities from all leafs the given rect intersects with within the tree.
//
// Retrieve returns a channel of entities. This is due to the fact that all Read-Only operations within
// Quadpix are run on there own thread.
func (q *Quadpix) Retrieve(rect pixel.Rect) <-chan Entities {
	out := make(chan Entities)

	go func() {
		out <- q.retrieve(rect)
		close(out)
	}()

	return out
}

// Intersect returns whether or not the given pixel.Rect intersects any entity with in the tree.
//
// Intersect returns a channel of a bool. This is due to the fact that all Read-Only operations within Quadpix are run on there own thread.
func (q *Quadpix) Intersect(rect pixel.Rect) <-chan bool {
	out := make(chan bool)

	go func() {
		out <- q.intersect(rect)
		close(out)
	}()

	return out
}

// Intersects returns all a channel of all entities that intersect with the given pixel.Rect within the tree.
//
// Intersects returns a channel of Entities due to the fact that all Read-Only operations in Quadpix are run on there own thread.
func (q *Quadpix) Intersects(rect pixel.Rect) <-chan Entities {
	out := make(chan Entities)

	go func() {
		out <- q.retrieve(rect).Intersects(rect)
		close(out)
	}()

	return out
}

// IsEntity returns whether or not the given entity exists with in the tree.
//
// IsEntity returns a channel holding a bool. This is due to the fact that all
// Read-Only operations in Quadpix are run on there own thread.
//
// Note that the given entity must have the same ID and pixel.Rect bounding box to be
// found as a match with in the tree.
func (q *Quadpix) IsEntity(entity *Entity) <-chan bool {
	out := make(chan bool)

	go func() {
		out <- q.isEntity(entity)
		close(out)
	}()

	return out
}

// tree node
type node struct {
	rect     pixel.Rect
	entities Entities
	children []*node
	depth    uint16
}

// create new node from given pixel.Rect bounds and prior nodes data.
func (n *node) new(rect pixel.Rect) *node {
	return &node{
		rect:     rect,
		entities: make(Entities, 0, cap(n.entities)),
		children: make([]*node, 0, 4),
		depth:    n.depth + 1,
	}
}

// recessive function for inserting entity's in to the tree.
func (n *node) insert(entity *Entity, maxDepth uint16) {
	// check for if you are at a leaf node.
	if len(n.children) > 0 {
		// find children the given entity's pixel.Rect intersects.
		nodes := n.getQuadrant(entity.Rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// recursive call to insert for each child node found.
		for i := range nodes {
			nodes[i].insert(entity, maxDepth)
		}
		return
	}

	// check for a needed split
	if len(n.entities)+1 > cap(n.entities) && n.depth < maxDepth {
		// split node in to its children
		n.split()

		// move this nodes entities to the new children nodes.
		n.moveEntities(append(n.entities, entity))
		return
	}

	// add entity to this nodes entities
	n.entities = append(n.entities, entity)
}

// split this nodes children in to there corresponding child quadrant nodes.
func (n *node) split() {
	n.children = append(n.children,
		n.new(pixel.R(n.rect.Min.X, n.rect.Min.Y, n.rect.Center().X, n.rect.Center().Y)),
		n.new(pixel.R(n.rect.Center().X, n.rect.Min.Y, n.rect.Max.X, n.rect.Center().Y)),
		n.new(pixel.R(n.rect.Min.X, n.rect.Center().Y, n.rect.Center().X, n.rect.Max.Y)),
		n.new(pixel.R(n.rect.Center().X, n.rect.Center().Y, n.rect.Max.X, n.rect.Max.Y)),
	)
}

// move given entities to this nodes children.
func (n *node) moveEntities(entities Entities) {
	for _, e := range entities {
		// find this entity's node
		nodes := n.getQuadrant(e.Rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// add this entity to each node found.
		for i := range nodes {
			nodes[i].entities = append(nodes[i].entities, e)
		}
	}

	// clear this nodes entities
	n.entities = n.entities[:0]
}

// remove the given entity from the tree.
//
// returns an error if no entity is found
func (n *node) remove(entity *Entity) error {
	// check for leaf
	if len(n.children) > 0 {
		// find nodes given entity intersects
		nodes := n.getQuadrant(entity.Rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// recursive remove call for each node found
		for i := range nodes {
			err := nodes[i].remove(entity)
			if err != nil {
				return err
			}
		}

		// attempted a collapse
		// does nothing of not needed
		n.collapse()

		return nil
	}

	// remove the given entity from the list of entities and retreave the new
	// list of entities with out the given entity
	entities, err := n.entities.Remove(entity)
	if err != nil {
		return err
	}

	// change this nodes entities to the new list of entities with out the given entity
	n.entities = entities

	return nil
}

// collapse collapses a node if the total number of entities from all child nodes is less then or
// equal to the max number of entities per node.
func (n *node) collapse() {
	// create a temp list of entities
	entities := make(Entities, 0)

	// attempted to merge all children entities in to the new entities list
	// this ignores all duplicate entities
	for i := range n.children {
		entities = entities.Merge(n.children[i].entities)
	}

	// check if the number of entities merged in to new list are less
	// then the cap of this nodes entities
	if len(entities) <= cap(n.entities) {
		// move found entities to this nodes entities
		n.entities = entities

		// remove children from this node
		n.children = n.children[:0]
	}
}

// retrieve gets all entities from all leafs the given pixel.Rect intersects
func (n *node) retrieve(rect pixel.Rect) (entities Entities) {
	// check for a leaf node
	if len(n.children) > 0 {
		// get all nodes pixel.Rect intersects
		nodes := n.getQuadrant(rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// recursive retrieve and merge call to add found entities to return list
		for i := range nodes {
			entities = entities.Merge(nodes[i].retrieve(rect))
		}
		return
	}

	// return this nodes entities
	return n.entities
}

// intersect checks if the given pixel.Rect intersects any entity with in the tree
func (n *node) intersect(rect pixel.Rect) bool {
	// check for a leaf
	if len(n.children) > 0 {
		// get all nodes the given pixel.Rect intersects
		nodes := n.getQuadrant(rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// check for intersects for all returned nodes
		for i := range nodes {
			if nodes[i].intersect(rect) {
				return true
			}
		}

		return false
	}

	// check for intersects with any entity with in this nodes entities
	return n.entities.Intersect(rect)
}

// isEntity checks if a given entity exists with in the tree
func (n *node) isEntity(entity *Entity) bool {
	// check if you are at a leaf
	if len(n.children) > 0 {
		// get all nodes the given entity intersects
		nodes := n.getQuadrant(entity.Rect)
		// panic error for no quadrent found.
		// this error should only occur if the given entity's bounds are not posable.
		// ie: the min and max are swapped in some way.
		if len(nodes) == 0 {
			panic(ErrNoNodeFound)
		}

		// recursive check for isEntity for all found nodes
		for i := range nodes {
			if nodes[i].isEntity(entity) {
				return true
			}
		}

		return false
	}

	// check if the nodes entities has the given entity
	return n.entities.Contains(entity)
}

// getQuadrant finds all nodes the given pixel.Rect intersects with
func (n *node) getQuadrant(rect pixel.Rect) (nodes []*node) {
	// check each child node for intersect
	for i := range n.children {
		// check if the child node rect intersects the given pixel.Rect
		if n.children[i].rect.Intersects(rect) {
			nodes = append(nodes, n.children[i])
		}
	}
	return
}
