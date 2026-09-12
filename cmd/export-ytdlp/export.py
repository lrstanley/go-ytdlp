#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///

import argparse
import json
import optparse
import re
import sys

from yt_dlp.extractor import list_extractor_classes
from yt_dlp.options import create_parser
from yt_dlp.version import CHANNEL, __version__


def option_default(option):
    default = option.default
    if default == optparse.NO_DEFAULT:
        return None
    if isinstance(default, (str, int, float, bool, list, dict, type(None))):
        return default
    try:
        return str(default)
    except Exception:
        return None


def option_type(option):
    if option.type == "choice":
        return "string"
    if option.type:
        return str(option.type)
    return "bool" if not option.takes_value() else "string"


def option_allows_multiple(option):
    if option.action == "append":
        return True
    if option.action != "callback":
        return False
    return bool((option.callback_kwargs or {}).get("append"))


def option_help(option):
    if option.help == optparse.SUPPRESS_HELP:
        return None
    if option.dest == "update_self" and "--update" in option._long_opts:
        return "Check if updates are available."
    if option.help and "%default" in option.help:
        return option.help.replace("%default", str(option_default(option)))
    if option.help:
        return re.sub(r"\s+\(Alias:.*\)\s*$", "", option.help)
    return option.help


def export_options(expected_version):
    if __version__ != expected_version:
        raise RuntimeError(
            f"yt-dlp version mismatch: expected {expected_version}, got {__version__}"
        )

    data = {
        "option_groups": [],
        "extractors": [
            {
                "name": ie.IE_NAME,
                "description": ie.description(markdown=False),
                "broken": not ie.working(),
                "age_limit": ie.age_limit or None,
            }
            for ie in list_extractor_classes()
        ],
        "channel": CHANNEL,
        "version": __version__,
    }

    parser = create_parser()
    for group in parser.option_groups:
        group_data = {
            "name": group.title,
            "description": group.description,
            "options": [],
        }
        for option in group.option_list:
            if option.dest == parser.ALIAS_DEST:
                continue

            option_data = {
                "id": option.dest,
                "action": str(option.action),
                "choices": list(option.choices) if option.choices else None,
                "help": option_help(option),
                "hidden": option.help == optparse.SUPPRESS_HELP,
                "meta_args": option.metavar,
                "type": option_type(option),
                "long_flags": option._long_opts,
                "short_flags": option._short_opts,
                "executable": option.action == "version",
                "allows_multiple": option_allows_multiple(option),
                "nargs": (
                    option.nargs
                    if option.nargs and option.nargs > 0 and option.takes_value()
                    else 0
                ),
                "default_value": option_default(option),
                "const_value": option.const,
            }

            if not option_data["id"] or option_data["id"] == "_":
                if option._long_opts:
                    option_data["id"] = option._long_opts[-1].lstrip("-")
                elif option.callback:
                    option_data["id"] = option.callback.__name__.lstrip("_")

            group_data["options"].append(option_data)
        data["option_groups"].append(group_data)

    return data


def main():
    argument_parser = argparse.ArgumentParser()
    argument_parser.add_argument("version")
    args = argument_parser.parse_args()
    json.dump(export_options(args.version), sys.stdout, indent=4)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
