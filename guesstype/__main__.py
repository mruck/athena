#
# This mini-utility shows basic usage of this library.
#
# One arg shows how it views types
#   python3 -m guesstype '{"numbers":"867 4678 23", "b64":"aGVsbG8K", "nested":"{\"k\":123}"}'
#
# Two or more args does the "autogenerate random values" thing.
#   python3 -m guesstype '{"numbers":"867 4678 23", "b64":"aGVsbG8K", "nested":"{\"k\":123}"}' '{"numbers":"231 132 323", "b64":"dGhlcmUK", "different":"blah"}'
#

from .easy import *

def main(argv):
  if len(argv) == 2:
    # Show the value of one arg.
    for k, v in decode_one(argv[1]).items():
      print(k, ':', repr(v))
  elif len(argv) > 2:
    # Treat each arg as an example value and guess a type.
    G = GuessType(argv[1:])

    # Use the type to generate random values.
    subvalues = G.indicator_values()
    for k, v in subvalues.items():
      print(k, ':', repr(v))
    print('----------')

    # Get the random value back into the input form
    example = G.unflatten(subvalues)
    print(example)
    return 0
  else:
    print("Usage: %s value_string" % argv[0])
    print("\tShows internal values of the string")
    print("Usage: %s value_string1 value_string2 ..." % argv[0])
    print("\tUses aggregate type guess to generate a randomized value.")
    return 1

if __name__ == "__main__":
  import sys
  sys.exit(main(sys.argv))
