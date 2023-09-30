# Examples

Sunbeam is language agnostic, so you can use any language you want to build your commands.

In order to be valid, scripts must:

1. write a valid manifest to stdout when called without arguments
2. handle every commands declared in their manifest
    - the command payload will be sent as the first argument to the script
