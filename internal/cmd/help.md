Launch k6 with a seamless extension user experience.

The launcher acts as a drop-in replacement for the `k6` command. For more convenient use, it is advisable to create an alias or shell script called `k6` for the launcher. The alias can be used in exactly the same way as the `k6` command, with the difference that it generates the real `k6` on the fly based on the extensions you want to use.

The launcher will always run the k6 test script with the appropriate k6 binary, which contains the extensions used by the script. In order to do this, it analyzes the script arguments of the "run" and "archive" subcommands, detects the extensions to be used and their version constraints.

Any k6 command can be used. Use the `help` command to list the available k6 commands.
