# mutualexclusion

We are using the token-ring algorithm

## How to get started!

1. Open a terminal in the root of the git repository.
1. Use the command ```go run .```
    1. Use the ```-oPort``` flag to set *ownPort* for the node.
    1. Use the ```-nPort``` flag to set *nextClientPort* for the next node in the logical ring.
    1. Use the ```-first``` flag (boolean) to set mark the **first** node in the logical ring.
1. Press enter in a terminal to request the *critical zone*.

## Copy && paste to start (at least 3 clients)

Each in a separate terminal located in the root of the git repository.

    1. go run . -oPort 8080 -nPort 8081 -first

    2. go run . -oPort 8081 -nPort 8082

    3. go run . -oPort 8082 -nPort 8080
