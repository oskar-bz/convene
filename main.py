import toml
import platform

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


def main():
    # check for existing file
    print("Hey")


if __name__ == "__main__":
    main()
