This is a go implementation of erlang's [gen_fsm](http://erlang.org/documentation/doc-6.1/lib/stdlib-2.1/doc/html/gen_fsm.html) 

# Why?
Erlang's `gen_fsm` provides a nice way of implementing finite state machines, which can be described as:<br>
`State(S) x Event(E) -> Actions(A), State(S')` <br>
There are a couple of things nice about this:
* Event handlers for different states are just functions with states as function names, and event data as inputs.
* Each instance of fsm is run as a separate process providing concurrency.

# How to implement?
* Implement the init function to set initial state<br>
``func (f *MyFsm) Init(args... interface{}) State``

* Implement event handlers for all states. An event handler looks like this<br> 
`func (f *MyFsm) State_Event(event_data) (State, Timeout)` <br>
where `State` is the next desired state, and timeout the amount of time in `time.Duration` to wait to send a timeout event if nothing happens.<br> 
For eg. a door in `Locked` state will transition to `Open`, when the right code is entered. This will look like this: <br>
```
func (d *Door) Locked_Button(code rune) (go_gen_fsm.State,time.Duration){
        if match(code){
            return "Open", 5 * time.Second
        }else{
            return "Locked", -1
        } 
    }
```
Here, the door will `open` after the correct is entered, and a timeout event will be sent after 5 seconds. To handle the timeout event, we should define a handler for that. This will look like this: <br>
```
func (d *Door) Open_Timeout() (go_gen_fsm.State){
        return "Locked"
    }
``` 
Now, the door will again be `Locked` after the timeout.

* Create an instance of your fsm and start `go_gen_fsm`<br>
```
    f := new(MyFsm)
    g := go_gen_fsm.Start(f, args...)
``` 
where `args` will be passed to the `Init` method to initialize the fsm.

* Finally, To send an event, use the `SendEvent` api provided by `GenFsm` (from the above step)<br>
`g.SendEvent(EventName, EventData...) `


# How it works?
Internally, `go_gen_fsm` starts a go routine when `Start` is invoked. This acts as an event loop listening to events, and uses reflection to invoke the appropriate handlers, thereby providing concurrency.

# Example
A full implementation of a door state machine described [here](http://erlang.org/documentation/doc-6.0/doc/design_principles/fsm.html) is implemented in the `sample` package.