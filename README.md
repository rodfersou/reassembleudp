## Initial Setup

1. **Install nix:**
   Follow the instructions for your operating system [here](https://nixos.org/download.html).

2. **Install direnv:**
    ```bash
    nix-env -i direnv
    ```

3. **Configure direnv:**
    Add the following line to your `~/.zshrc`:
    ```bash
    eval "$(direnv hook zsh)"
    ```

4. **Copy .env example configuration:**
    ```bash
    cp docs/dotenv.example .env
    ```

5. **Allow direnv to open the shell in the project:**
    ```bash
    direnv allow
    ```

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

### Limitations
Understanding the task requirements, especially regarding fault tolerance, posed challenges.

It's saying that should be fault tolerant, but as far as I know this means that need to have multiple instances, if one fall other should take the request.

It sounds like a infrastructure problem, how to deploy safetly, Blue green deployment.

The application currently runs in a single execution, and restarting it recreates the database.

If you run twice the emitter, will get duplicate errors.  My initial tought about this was to clean the table after processing, but running both database and application in my machine got some instability, and database deletion was making things worse.

I opted to keep it working in 1 run to get an average of +70% success of message reassemble.

## How to Run

1. **Start MongoDB:**
    ```bash
    ./scripts/start_mongodb.sh
    ```

2. **Start the reassembleudp:**
    ```bash
    ./scripts/run_reassembleudp.sh
    ```

3. **Start UDP emitter:**
    ```bash
    ./scripts/run_emitter.sh
    ```

## How to Run Tests

1. **Run unit tests:**
    ```bash
    ./scripts/test_unit.sh
    ```

2. **Run end-to-end tests:**
    ```bash
    ./scripts/test_e2e.sh
    ```
