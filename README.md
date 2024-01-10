## Initial Setup

1. **Install docker:**
   Follow the instructions for your operating system [here](https://docs.docker.com/engine/install/).

## App Design

The application utilizes MongoDB as its database with two collections:

### Database

* **Fragments collection:**
  - Stores received fragments after parsing the binary and handles duplications.
* **Messages collection:**
  - Keeps track of message processing status and timeouts.
  - Message status can be one of: IN_PROGRESS, VALID, INVALID.

### Internal Structure

Upon startup, the application establishes connections to a UDP socket and MongoDB, shared among all resources. A thread pool runs four goroutines for reading, parsing, and saving fragments, including updating the message table. This pool performs bulk inserts of 4000 items to the database.

To process messages, an additional goroutine is responsible for reassembling the original message and checksumming it. On average, the application achieves over 70% success in this process.

```
+---------------------+           +---------------------+
|      UDP Socket     |           |      MongoDB        |
|                     |           |                     |
|   +--------------+  |           |   +-------------+   |
|   | Goroutine 1  |  |           |   | Fragments   |   |
|   +--------------+  |           |   | Collection  |   |
|   +--------------+  |           |   +-------------+   |
|   | Goroutine 2  |  |           |   +-------------+   |
|   +--------------+  |           |   | Messages    |   |
|   +--------------+  |           |   | Collection  |   |
|   | Goroutine 3  |  |           |   +-------------+   |
|   +--------------+  |           +---------------------+
|   +--------------+  |
|   | Goroutine 4  |  |
|   +--------------+  |
+---------------------+
   |               |
   +---------------+
```

### Limitations
Understanding the task requirements, especially regarding fault tolerance, posed challenges.

It's saying that should be fault tolerant, but as far as I know this means that need to have multiple instances, if one fall other should take the request.

It sounds like a infrastructure problem, how to deploy safetly, Blue green deployment.

The application currently runs in a single execution, and restarting it recreates the database.

If you run twice the emitter, will get duplicate errors.  My initial tought about this was to clean the table after processing, but running both database and application in my machine got some instability, and database deletion was making things worse.

I opted to keep it working in 1 run to get an average of +70% success of message reassemble.

### Known issues
Another discovery is that MongoDB create indexes 4 times faster if run in foreground instead of in normal background execution, according to [this article](https://medium.com/@KarakasP/time-difference-between-background-and-foreground-index-creation-in-mongodb-b29ca3689fdc)

## How to Run

1. **Run the app with Docker:**
    ```bash
    ./scripts/docker_run.sh
    ```

## How to Run Tests

1. **Run unit tests:**
    ```bash
    ./scripts/docker_test_unit.sh
    ```

2. **Run end-to-end tests:**
    ```bash
    ./scripts/docker_test_e2e.sh
    ```
