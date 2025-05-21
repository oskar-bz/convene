import toml
import platform
import os
import sys

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


CONFIG_PATH

def get_script_path():
    return os.path.dirname(os.path.realpath(sys.argv[0]))

def get_config():
    with open(get_script_path() + "/config.toml")

def main():
    # check for existing file
    config = get_config()
        
    save_config()

if __name__ == "__main__":
    main()
