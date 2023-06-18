# SyncSaga

SyncSaga is a consensus mechanism designed for game engines, aiming to achieve synchronization of game states among multiple players and maintain the coherence of the game's storyline. This project focuses on resolving synchronization issues in game development to ensure consistent and conflict-free gaming experiences among players.

## ReadyGroup Usage

ReadyGroup is a synchronization mechanism designed for waiting until all game players are ready. It resembles the concept of sync.WaitGroup, but provides more flexibility and callbacks to manage the state of the game players.

Here is the basic usage of ReadyGroup. With this tool, we can manage and synchronize the state of game players more conveniently.

### Create a ReadyGroup

To use ReadyGroup, you first need to create a new instance. Here is an example of how to create a ReadyGroup:

```golang
rg := NewReadyGroup(
    WithTimeout(0, nil),
    WithUpdatedCallback(func(rg *ReadyGroup) {
        // Callback when state is updated
    }),
    WithCompletedCallback(func(rg *ReadyGroup) {
        // Callback when all players are ready
    }),
)
```

The NewReadyGroup method accepts four parameters:

* `WithTimeout(interval int, callback ReadyGroupCallback)` - Sets a timeout period (in seconds). If a callback function is provided, it will be called when the timeout occurs.
* `WithValidator(v ReadyGroupValidator)` - Provides a validation function that is called every time a state update occurs. It should return a boolean value indicating whether all participants are ready. If true is returned, the completed callback is triggered; if false is returned, it is not.
* `WithUpdatedCallback(callback ReadyGroupCallback)` - This function is called whenever any participant's state is updated.
* `WithCompletedCallback(callback ReadyGroupCallback)` - This function is called when all participants are ready.

### Adding Participants

To add participants to the ReadyGroup, you can do:

```golang
rg.Add(0, false)
rg.Add(1, false)
rg.Add(2, false)
rg.Add(3, false)
```

Here, the Add method takes two parameters: the id of the participant and their initial state. In this example, we added 4 participants with ids ranging from 0 to 3 and set their initial state to false.

### Starting the Wait

After adding participants, you can call rg.Start() to start waiting for the participants to become ready.

```golang
rg.Start()
```

### Updating Participant States

When participants become ready, you can call the rg.Ready(id int64) method to update their state, where id is the id of the participant. Here's how you can update all participant states:

```golang
rg.Ready(0)
rg.Ready(1)
rg.Ready(2)
rg.Ready(3)
```

In this example, we used a goroutine to update all participants' states, but you can call the rg.Ready(id int64) method as needed.

Note: The state update operations should be performed after calling rg.Start(), or the callbacks may not be triggered correctly.
