import toml
import platform
import os
import sys
import subprocess
import asyncio

# MYDOTS
# Commands:
# - sync:
#     Gathers all config files and pushes the current config to the git repository, requires git and an upstream repository
#     also compares the ```generation``` of the upstream and downstream configs to prevent conflicts (maybe compare hashes)
# - pull:
#     replaces the current config files with the contents of the upstream repository
# - add:
#     adds the config of an application to be tracked, requires a path (and optional other paths for other OSes) and an application name
# - remove:
#     stops tracking an application


def get_script_path():
    return os.path.dirname(os.path.realpath(sys.argv[0]))


def get_config():
    if sys.platform == "win32":
        config_path = "%appdata%/../local/convene/"
    elif sys.platform == "linux" or sys.platform == "linux2":
        config_path = "/etc/convene/"
    else:
        print("ERROR:", sys.platform, "is not a supported operating system")

    if not os.path.exists(config_path + "config.toml"):
        print("No config found")
        return None

    with open(get_script_path() + "/config.toml") as f:
        config = toml.loads(f.read())
        return config


def save_config():
    pass


def run(cmd):
    proc = subprocess.Popen(
        cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
    )

    return proc.stdout


def main():
    stdout = run("type main.py && timeout /t 5 && type pyproject.toml")
    if stdout == None:
        print("stdout is none!")
        return
    line = stdout.readline().decode()
    lines = [line]
    i = 0
    outputs = ["a", "b", "c"]
    while line:
        # print("\rReading" + "." * (i + 1), end="")
        print("\r" + outputs[i] + str(i), end="")
        line = stdout.readline()
        lines.append(line.decode())
        i = (i + 1) % 3

    return

    # check for existing file
    config = get_config()
    # check for git
    try:
        out = subprocess.check_output(["git", "--version"])
    except:
        print(
            "Convene needs git to synchronize your files. Make sure it is installed and in your PATH."
        )
        exit(-1)

    if config == None:
        config = {}
        print(
            "Welcome to Convene!\nTo start syncing your files, you need to create a repository first."
        )
        config["link"] = input("Repository-link: ")
        # clone the repo
        print("Cloning")
    save_config()


if __name__ == "__main__":
    main()
