import pytest
import subprocess
import time


def run_command(command):
    process = subprocess.Popen(
        command.split(), stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, shell=False
    )
    while True:
        output = process.stdout.readline()
        if output == "" and process.poll() is not None:
            break
        if output:
            yield output.decode().strip()
    rc = process.poll()
    return rc


def test_e2e():
    reassemble_output = run_command("go run reassembleudp.go")
    emitter_output = run_command("node udp_emitter.js")

    emitter_messages = {}
    reassemble_messages = {}
    while len(set(emitter_messages) & set(reassemble_messages)) < 10:
        msg = next(emitter_output)
        message_id = int(msg.split()[2][1:])
        data_size = int(msg.split()[4].split(":")[1])
        checksum = msg.split()[6].split(":")[1]
        emitter_messages[message_id] = (data_size, checksum)

        msg = next(reassemble_output)
        while "Hole" in msg:
            msg = next(reassemble_output)
        message_id = int(msg.split()[1][1:])
        data_size = int(msg.split()[3])
        checksum = msg.split()[4].split(":")[1]
        reassemble_messages[message_id] = (data_size, checksum)

    for message_id in set(emitter_messages) & set(reassemble_messages):
        assert emitter_messages[message_id] == reassemble_messages[message_id]
