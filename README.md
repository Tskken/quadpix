# Quadpix [![GoDoc](https://godoc.org/github.com/Tskken/quadpix?status.svg)](https://godoc.org/github.com/Tskken/quadpix) [![Go Report Card](https://goreportcard.com/badge/github.com/Tskken/Quadpix)](https://goreportcard.com/report/github.com/Tskken/Quadpix) [![Build Status](https://travis-ci.org/Tskken/quadpix.svg?branch=master)](https://travis-ci.org/Tskken/quadpix)
 
Quadpix is a quad-tree implementation spawned from my other QuadGo library but designed to directly intigrate with [pixel](https://github.com/faiface/pixel).
As this library is a work off of the existing QuadGo library much of the structure is the same but a few things have been changed to fallow some of the pixel naming conventions and styles.
 
# Getting Started
To get Quadpix, run `go get github.com/Tskken/quadpix` in your command line of choice.
Then add it to any of your existing projects by adding `import "github.com/Tskken/quadpix"`.

# Requirements
You must have `github.com/faiface/pixel` for this library to work. If you are using go modules it should be installed for you if you done already have it.
Note that you must have one of the most up to date versions of pixel as of creations of this ReadMe. That means if you have a prior installation of pixel and you have not
updated it in some time, ie at least not up to some of the most reascent pull request for v0.8.1 as of Nov, 5, 2019, Quadpix will not work as it relies on some newer features added in some of the most reascent updates.
 
# Tutorial
 
## Creating the quad-tree
 
To create your instance of Quadpix you have to call the New() function. This will create a new instance of Quadpix and return its reference to be used.
 
For example:
 
```go
    // create a basic instance
    tree := quadpix.New(width, height, maxEntities, maxDepth)
```
 
## Adding entities to the tree
 
The main way to insert any data in to the quadpix tree is through the Insert function. This function takes the pixel.Rect bounds of what ever entity you want to insert and a variadic number of Action functions. I will talk about Action functions next but first to insert a basic pixel.Rect with a min of 0, 0 and a max of 50, 50, you would do as shown below.
 
Example:
```go
    // insert an entity with a bounds of min:0, 0, max: 50, 50
    tree.Insert(pixel.R(0, 0, 50, 50))
```
 
This function as showned will create a new quadpix.Entity type with the given pixel.Rect bounds. Check the Godoc's for more info on what a quadpix.Entity is and what kind of data is holds.

As stated above there is also a thing called an Action which can be given to Insert. The function signature of Insert is as shown below.
```go
    quadpix.Insert(pixel.Rect, Action...)
```
The variadic argument Action takes any number of Action functions to be saved to the new quadpix.Entity that is created. This Action functions are simply just empty functions with no signature as so `func()`. They can be treated as lambda functions to be saved and possibly used later for running code on intersect with the entity.
 
An example of adding an Action function on insert would be as shown below.
```go
    // Insert with an Action function
    tree.Insert(pixel.R(0, 0, 50, 50), func(){
        // some action done that is stored in Entity
    })
```
 
Another way to insert any number of Entity's in to the tree is through InsertEntities(). This function takes a variadic number of Entity's and will insert all of them in to the tree.  For example:
```go
    // Insert one entity in to the tree through InsertEntity
    tree.InsertEntities(quadpix.E(pixel.R(0, 0, 50, 50)))

    // Create a list of entities to insert
    entities := ...

    // Insert a list of entities in to the tree
    tree.InsertEntities(entities...)
```
 
## Removing entities from the tree
 
To remove entities from the tree you need to use quadpix.Remove(). This function will remove the given entity from the tree and if needed collapse all nodes to save memory space and clean up the tree.
 
Example:
```go
    // remove the entity from the tree
    err := tree.Remove(entity)
    if err != nil {
        panic(err)
    }
```
 
If you note there is a return type of error on Remove(). If the given entity is not found within the tree, Remove() will return an ErrNoEntityFound.
 
One important thing to understand is that the given entity to Remove() has to at least have the same ID and Bound as the entity you want to remove. This is because when trying to find the entity to remove it uses the entity’s ID and Bound to check if the entity found is the one you want to remove.
 
#### Retrieving entities from the tree
 
To find entities in the tree you need to use quadpix.Retrieve(). This function takes a pixel.Rect to search the tree with and will return all entities from nodes that that given pixel.Rect intersects with.
 
Example:
```go
    // get all entities from nodes that pixel.Rect intersects
    entities := <-tree.Retrieve(pixel.R(0, 0, 50, 50))
```
 
As a note, if you look at the `<-tree ...` part of the code, this is because Retrieve() returns a channel of Entities not just Entities. This is because all Read-Only operations in Quadpix are run concurrently. In the case above we are calling tree.Retreive() and blocking till some amount of entities are returned on the output channel of Entities. We then assign that value to entities to use. Because of the fact that this function is concurrent, we can also run the Retrieve() function and receive from the channel later similar to how a time.Tick() works in the Go standard library.
 
Example:
```go
    // get all entities for the given bounds but not receive
    // from the channel till later.
    out := tree.Retrieve(pixel.R(0, 0, 50, 50))
 
    // do some other stuff that doesn't need the value from Retrieve()
    ...
 
    // get the entities from the channel or block if needed
    entities := <-out
```
 
As you will see in the next section's, and as stated earlier, all Read-Only operations for the tree are run concurrently. This means all read operations return a similar structure of a channel type to be used however is best for that instance.
 
Also to note data of any kind will only ever be sent to the channel once and then the channel is closed. This means that if you try and receive from the channel two times the code will given an error for trying to retreave on a closed channel.
 
Example:
```go
    // get entities for bounds
    out := tree.Retrieve(bound)
 
    // do stuff
    ...
 
    // receive from channel
    entities := <-out
 
    // do more stuff
    ...
 
    // error due to a second receive call from a closed channel
    entities2 := <-out
```
 
 
Another important thing to note is that because these operations are run concurrently, and because Quadpix does not have any native syncing functionality. This means that even though read operations are run concurrently, all write actions are not and could change data while a read operation is executing. This means you have to make sure that you are not using Insert() or Remove() well one of the read operations are running or you may get some unexpected behavior. Due to this it is a best practice to make sure to ether insert before you run any read operations or block from your read operations to before doing any insert or remove operations to ensure no data is being changed when it is not suppost to.
 
 
## Checking for collisions
 
One of the most important parts of a quad-tree made for collision detection in video games is the functions used to check for collisions. This takes the form of two functions with in Quadpix, the Intersect() and Intersects() functions. As the names kind of imply the Intersect() function checks if a given bound intersects any entity with in the tree. The Intersects() function returns all entities that the given bound intersects within the tree.
 
Example:
```go
    // check for collision with a bounds
    if <-tree.Intersect(pixel.R(0, 0, 50, 50)) {
        // do something on intersect case
        ...
    }
 
    // get all entities the bounds intersects with
    entities := <-tree.Intersects(pixel.R(0, 0, 50, 50))
```
 
As you can see, and as said in the prior section, the collision check functions are also considered Read-Only functions and in turn run concurrently. This means they also return channels and can be used the same way Retrieve() was used up above.
 
Additional these functions run in to the same possible issues as Retrieve() as they only ever can receive from the channel once and are not safe to run concurrently with Insert() or Remove().
 
## Other useful functions
 
There is one other possibly useful function provided by Quadpix. This is the IsEntity() function. This function checks to see if the given entity exists with in the tree. Similarly with Remove() the given entity has to have the same ID and pixel.Rect as the entity you are trying to find. This could be useful if you want to check to make sure an entity was removed from the tree or to check to see if an entity exists with in the tree and if not add it back.
 
Example:
```go
    // check if the entity exists with in the tree
    if !<-tree.IsEntity(entity) {
        // do something if entity didn't exist
        ...
    }
```
 
This function is also a Read-Only function so its run concurrently. So make sure if you are using it to not run Insert() or Remove() at the same time.
 
# Feature requests and bug reports
 
If you have any ideas for new features or find any bugs with this library please make an issue report and I will get to it as soon as I can.
 
# Current TODO
- Re-write benchmarks for the library.
- More tests for possible edge cases.
 
## License
 
[MIT](LICENSE)
