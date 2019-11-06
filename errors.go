package quadpix

import "errors"

var (
	// ErrNoEntityFound error
	ErrNoEntityFound = errors.New("no entity found")

	// ErrNoNodeFound error
	ErrNoNodeFound = errors.New("no node found from getQuadrent()")

	// ErrNoEntitiesGiven error
	ErrNoEntitiesGiven = errors.New("no entities given to InsertEntities()")
)
