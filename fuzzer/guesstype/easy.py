from .stats import AnyStats, default_pathprops
import random

class GuessType:
  def __init__(self, examples):
    """Initialize with examples of the thing of indeterminite type."""
    self.stats = AnyStats()
    assert not isinstance(examples, (str, bytes)), "you probably didn't want this"
    for i in examples:
      self.stats.add(i)

  def flatten_stats(self):
    """Returns a dict.
    The keys are 'virtual paths', indicating some sub-component of the input.
    A virtual path is a tuple meant to be treated opaquely, with one exception:
    if the virual path contains an python Ellipsis object, the k/v pair represents
    an optional key that may be omitted (e.g. when you unflatten).
    The values are "LeafStats" instances, which tell you about the type.
    """
    t = self.stats.paths(props=default_pathprops._replace(mono=True))
    assert len(t) != 0, "no examples given!"
    assert len(t) == 1, "oops"
    return t[0]

  def flatten_stats_many(self):
    """Like flatten, except returns a list of dicts.
    Each dict indicates a separate possibility."""
    return self.stats.paths(props=default_pathprops._replace(mono=False))

  def unflatten(self, pathvals):
    """Takes a dict from flatten_stats(), except the values should be
    real python objects (like strings and ints) instead of LeafStats
    instances.
    Basically, to use this you iterate through a .flatten_stats() dict
    and change all the values (and none of the keys).
    """
    return self.stats.from_pathvals(pathvals)

  def indicator_values(self, *, R=None, discard_optional=False):
    """Returns a dict of 'virtual paths' keys to indicator sub-values within
    the generated object. The result may be passed to unflatten to get the
    encoding of the dict. You can modify the values (to something of the same
    type) if you wish."""
    if R is None:
      R = random.Random()

    flat = self.flatten_stats()
    out = {}
    for k in flat:
      if discard_optional and Ellipsis in k:
        continue
      out[k] = flat[k].gen_ident(R)

    return out

  def indicator_values_many(self, *, R=None, discard_optional=False):
    """Like .indicator_values(), but returns many possibilities where
    things that appear to be union types are different and such."""
    if R is None:
      R = random.Random()

    flats = self.flatten_stats_many()
    for flat in flats:
      out = {}
      for k in flat:
        if discard_optional and Ellipsis in k:
          continue
        out[k] = flat[k].gen_ident(R)

      yield self.unflatten(flat), flat

def decode_one(example):
  stats = AnyStats()
  stats.add(example)
  d = stats.paths(props=default_pathprops._replace(mono=True))[0]
  for k in d:
    d[k] = d[k].record[0]
  return d
