Todos
=====

* The query update tests simply allow 409, even if it is the first
time. In other tests we have distinguished this, which might be
better. But this is a niceness at best.
* The query update tests forbid a response 404 and return an error if
this happens. This might bite us eventually, as far as I see. If chaos
delays the update query for too long, the cursor or a piece of it
could time out in the cluster and then the cursor is simply gone and
respond with 404. This is - in a sense - allowed behaviour, so this
could lead to a false positive. However, I suggest we let it run for
now. This situation can be identified easily, when it happens.
* We should additionally test streaming cursors. Potentially, we can
simply flip a coin and do some of the tests with streaming if the coin
comes up heads. This would then make test coverage a bit more
difficult. So maybe we just want to add some tests with streaming
cursors. But this is also for later.
