# TODO: Stats.decode method, to force a guess onto a concrete example.
#   (kinda like easy.decode_one, but with a forced guess rather than free.
#    Also, with potentially multiple guesses).
# Bugs: bytes aren't particularly handled well; use strings everywhere.

from collections import namedtuple
from itertools import product
import datetime
from .guessstring import *
from . import guesskeygroups

# 1990 to 2040
EPOCH_RANGE = (631152000, 2208988800)

# Sometimes used to adjust randomness for gen_ident.
APPROX_IDENT_BITS = 10

#
# Helpers
#

class PathProperties(namedtuple('PathProperties', 'mono')):
  __slots__ = ()

# non-mono means .paths() may return multiple sets of paths.
# There's a bit of a combinatorial explosion issue, so, don't set it
# for types that may turn out to be complex. (Maybe PathProperties
# should come with a 'gimme 20' instead of all-or-one).
default_pathprops = PathProperties(mono=True)

#
# Stats objects aggregate information about the values seen so far.
#

class LeafStats:
  """Base class and default implementations for leaf types"""
  def __init__(self):
    self.total = 0
    self.record = []

  def add(self, val):
    self.total += 1
    self.record.append(val)

  def paths(self, *, prefix=(), props=default_pathprops):
    return [{prefix: self}]

  def from_pathvals(self, pathvals):
    assert len(pathvals) == 1, "Leaves can only handle one path. %r" % (pathvals,)
    k, v = list(pathvals.items())[0]
    assert k == (), "Leaves can only handle empty path. %r" % (k,)
    return v

  def gen_ident(self, R):
    """gen_ident is for generating a random value that lets us identify it among others.
    gen_ident is only implemented for leaf values."""
    raise NotImplementedError("missing gen_ident")

class NumStats(LeafStats):
  # TODO: Doesn't handle NaNs well.
  def __init__(self):
    super().__init__()
    self.min = float('inf')
    self.min_positive = float('inf')
    self.max = float('-inf')
    self.float_count = 0
    self.int_count = 0
    self.unix_date_count = 0
    self.unix_milli_count = 0
    self.zeros_count = 0

  def add(self, val):
    super().add(val)
    self.min = min(self.min, val)
    self.max = max(self.max, val)
    if val > 0:
      self.min_positive = min(self.min_positive, val)

    if EPOCH_RANGE[0] <= val <= EPOCH_RANGE[1]:
      self.unix_date_count += 1
    elif EPOCH_RANGE[0] <= val*1000 <= EPOCH_RANGE[1]:
      self.unix_milli_count += 1

    if isinstance(val, int):
      self.int_count += 1
    elif isinstance(val, float):
      self.float_count += 1

    if val == 0:
      self.zeros_count += 1

  @property
  def likely_date(self):
    dates = self.unix_date_count + self.unix_milli_count
    nonzeros = self.total - self.zeros_count
    return dates >= 1 and dates > nonzeros * 0.8

  @property
  def likely_float(self):
    nonzeros = self.total - self.zeros_count
    return self.float_count >= 1 and self.float_count > nonzeros * 0.1

  def gen_ident(self, R):
    infs = (float('inf'), float('-inf'))
    minv = self.min if self.min not in infs else 0
    maxv = self.max if self.max not in infs else 100
    if self.likely_float:
      return float('%.03f' % R.uniform(float(minv), float(maxv)))
    else:
      maxv, minv = int(maxv), int(minv)
      if maxv - minv < (1<<APPROX_IDENT_BITS):
        maxv = minv + (1<<APPROX_IDENT_BITS)
      return R.randint(minv, maxv)

class BoolStats(LeafStats):
  def gen_ident(self, R):
    # There's no actual way to have enough entropy. Whatever.
    return R.randint(0, 1) == 1

class DateStats(LeafStats):
  def gen_ident(self, R):
    return datetime.datetime.now() + datetime.timedelta(seconds=R.randrange(0, (1<<APPROX_IDENT_BITS)))

class NullStats(LeafStats):
  def gen_ident(self, R):
    return None

class MissingStats(LeafStats):
  def gen_ident(self, R):
    return Ellipsis # This should probably get ignored.

class BasicStrStats(LeafStats):
  def __init__(self):
    super().__init__()
    self.min_len = float('inf')
    self.min_len_nonzero = float('inf')
    self.max_len = float('-inf')
    self.charset = set()

  def add(self, val):
    assert isinstance(val, (str, bytes))
    super().add(val)
    self.min_len = min(self.min_len, len(val))
    self.max_len = max(self.max_len, len(val))
    if len(val) > 0:
      self.min_len_nonzero = min(self.min_len_nonzero, len(val))
    self.charset.update(val)

  def gen_ident(self, R):
    # There's no actual way to have enough entropy... whatever.
    length = min(max(0, self.max_len), 32)
    charset = [c for c in self.charset if 32 < ord(c) < 127]
    try:
      big_enough = len(charset) ** length >= (1<<APPROX_IDENT_BITS)
    except OverflowError:
      big_enough = True

    if not big_enough:
      return ''.join(R.choice('qwertyuiopasdfghjklzxcvbnm') for i in range(max(length, 8)))
    else:
      return ''.join(R.choice(charset) for i in range(length))

class StrStats:
  def __init__(self):
    self.total = 0
    self.min_len = float('inf')
    self.max_len = float('-inf')
    self.guess_counts = Counter()
    self.guess_info = {}

  def add(self, s):
    if isinstance(s, StringGuesser):
      guesser = s
      s = s.s
    else:
      guesser = StringGuesser(s)

    self.total += 1
    self.min_len = min(self.min_len, len(s))
    self.max_len = max(self.max_len, len(s))

    st, val = guesser.best
    self.guess_counts[st] += 1
    if isinstance(st, StrJSON):
      if st.type == 'dict':
        self.guess_info.setdefault(st, DictStats()).add(val)
      elif st.type == 'list':
        self.guess_info.setdefault(st, SeqStats()).add(val)
      else:
        assert False, "Only dict and list json types are handled."
    elif isinstance(st, StrHex):
      if isinstance(val, bytes):
        try:
          val = val.decode('utf-8')
        except UnicodeDecodeError:
          val = val.decode('ISO-8859-1') # meh, whatever, we're bad at bytes anyways.
          return self.guess_info.setdefault(st, BasicStrStats()).add(val)
      self.guess_info.setdefault(st, StrStats()).add(val)
    elif isinstance(st, StrB64):
      if isinstance(val, bytes):
        try:
          val = val.decode('utf-8')
        except UnicodeDecodeError:
          val = val.decode('ISO-8859-1') # meh, whatever, we're bad at bytes anyways.
          return self.guess_info.setdefault(st, BasicStrStats()).add(val)
      self.guess_info.setdefault(st, StrStats()).add(val)
    elif isinstance(st, StrNum):
      self.guess_info.setdefault(st, NumStats()).add(val)
    elif isinstance(st, StrList):
      self.guess_info.setdefault(st, SeqStats()).add(val)
    elif isinstance(st, StrBool):
      self.guess_info.setdefault(st, BoolStats()).add(val)
    elif isinstance(st, StrDate):
      self.guess_info.setdefault(st, DateStats()).add(val)
    elif isinstance(st, StrStr):
      self.guess_info.setdefault(st, BasicStrStats()).add(val)
    else:
      assert False, "Unknown string guess"

  def best_type(self):
    return self.guess_counts.most_common(1)[0][0]

  def paths(self, *, prefix=(), props=default_pathprops):
    if props.mono:
      types = [self.best_type()]
    else:
      types = self.guess_info.keys()

    res = []
    for k in types:
      res.extend(self.guess_info[k].paths(prefix=prefix+(k,), props=props))
    return res

  def from_pathvals(self, pathvals):
    # Handle special case of BasicStr-less string first.
    if len(pathvals) == 1:
      k, v = list(pathvals.items())[0]
      if k == ():
        return v

    assert len(pathvals) > 0, "Can't make str from empty pathvals"
    akey = next(iter(pathvals.keys()))
    assert len(akey) > 0, "Can't make str from empty keys"
    st = akey[0]
    assert isinstance(st, StrTypeBase), "Str paths must be a StrTypeBase"
    subpaths = {}
    for k, v in pathvals.items():
      assert k[:1] == (st,), "Str path StrTypeBases must match. %r != %r" % ((st,), k[:1])
      subpaths[k[1:]] = v
    return st.invert(self.guess_info[st].from_pathvals(subpaths))

class AnyStats:
  def __init__(self):
    self.total = 0
    self.stats = {}

  def copy(self):
    new = AnyStats()
    new.total = self.total
    new.stats = self.stats.copy()
    return new

  def add_missing(self):
    self.total += 1
    self.stats.setdefault('missing', MissingStats()).add(None)

  def add(self, obj):
    self.total += 1
    if isinstance(obj, dict):
      self.stats.setdefault('dict', DictStats()).add(obj)
    elif isinstance(obj, (list, tuple)):
      self.stats.setdefault('seq', SeqStats()).add(obj)
    elif isinstance(obj, (str, bytes, StringGuesser)):
      self.stats.setdefault('str', StrStats()).add(obj)
    elif isinstance(obj, bool):
      self.stats.setdefault('bool', BoolStats()).add(obj)
    elif isinstance(obj, (int, float)):
      self.stats.setdefault('num', NumStats()).add(obj)
    elif obj is None:
      self.stats.setdefault('null', NullStats()).add(obj)
    else:
      assert False, "Unhandled object type"

  def best_type(self):
    if self.total == 0:
      return 'missing'
    return max((v.total, k) for k, v in self.stats.items())[1]

  def good_types(self):
    if self.total == 0:
      return ['missing']
    return [k for k, v in self.stats.items() if v.total >= self.total * 0.1]

  def paths(self, *, prefix=(), props=default_pathprops):
    if props.mono:
      types = [self.best_type()]
    else:
      types = self.good_types()

    res = []
    for k in types:
      v = self.stats[k]
      res.extend(v.paths(prefix=prefix+(k,), props=props))
    return res

  def from_pathvals(self, pathvals):
    assert len(pathvals) > 0, "Can't make any from empty pathvals"
    akey = next(iter(pathvals.keys()))
    assert len(akey) > 0, "Can't make any from empty keys"
    typ = akey[0]
    assert typ in ('dict', 'seq', 'str', 'bool', 'num', 'null'), "Any has unknown tag."
    subpaths = {}
    for k, v in pathvals.items():
      assert k[:1] == (typ,), "Any path tags must match. %r != %r" % ((typ,), k[:1])
      subpaths[k[1:]] = v
    return self.stats[typ].from_pathvals(subpaths)

class DictStats:
  # Keys are assumed to be only strings or ints.
  # Doesn't currently do any inference on keys.
  # Consider detecting "ordered" keys like "obj1", "obj2", ...
  def __init__(self):
    self.total = 0
    self.items = {}
    self.always_keys = None
    self.alone_keys = set()

    # For each key, sibling_keys is a set of keys that have been seen
    # to appear with that key. And spousal_keys is a set of keys that
    # have *always* been seen to appear with that key.
    self.sibling_keys = {}
    self.spousal_keys = {}

  def add(self, d):
    self.total += 1
    d_keys = d.keys()

    if self.always_keys is None:
      self.always_keys = set(d_keys)
    else:
      self.always_keys.intersection_update(d_keys)

    if len(d_keys) == 1:
      self.alone_keys.add(list(d_keys)[0])

    # Note: we don't add 'missing' fields.
    for k, v in d.items():
      if k not in self.items:
        t = AnyStats()
        t.add(v)
        self.items[k] = t
        self.sibling_keys[k] = set(d_keys)
        self.spousal_keys[k] = set(d_keys)
      else:
        self.items[k].add(v)
        self.sibling_keys[k].update(d_keys)
        self.spousal_keys[k].intersection_update(d_keys)

  def guess_key_groups_mo(self):
    """A heuristic for representing allowed groups of keys.
    It assumes that each group has a set of mandatory keys,
    and a set of optional keys. (This is not necessarily true,
    and this is not the only assumption this heuristic makes).
    """
    return guesskeygroups.guess_key_groups_mo(self.sibling_keys, self.spousal_keys, self.alone_keys)

  def paths(self, *, prefix=(), props=default_pathprops):
    if props.mono:
      groups = [(frozenset(self.always_keys), frozenset(self.items.keys() - self.always_keys))]
    else:
      groups = self.guess_key_groups_mo()

    memo = {}
    def subpath(k, optional=False):
      memo_k = (k, optional)
      try:
        return memo[memo_k]
      except KeyError:
        pass
      v = self.items[k]
      # "Ellipsis" is used as a "optional" sentinel; any path with Ellipsis
      # in it is apparently optional.
      p = (k,) if not optional else (Ellipsis, k)
      ret = v.paths(prefix=prefix+p, props=props)
      memo[memo_k] = ret
      return ret

    whole_res = []
    for mandatory, optional in groups:
      group_res = []
      for k in mandatory:
        group_res.append(subpath(k))
      for k in optional:
        group_res.append(subpath(k, optional=True))
      for combo in product(*group_res):
        d = combo[0]
        for i in range(1, len(combo)):
          d.update(combo[i])
        whole_res.append(d)
    return whole_res

  def from_pathvals(self, pathvals):
    res = {}
    for k, v in pathvals.items():
      if k[:1] == (Ellipsis,):
        # Ellipsis path component is ignored; if the didn't want it
        # they should have deleted it!
        assert len(k) >= 2, "dict key too short"
        res.setdefault(k[1], {})[k[2:]] = v
      else:
        assert len(k) >= 1, "dict key too short"
        res.setdefault(k[0], {})[k[1:]] = v

    for k in res:
      res[k] = self.items[k].from_pathvals(res[k])
    return res

class SeqStats:
  def __init__(self):
    self.total = 0
    self.min_len = float('inf')
    self.max_len = float('-inf')
    self.zero_len = 0
    # Track stats separatley for the first 32 objects.
    self.initial_stats = [AnyStats() for i in range(32)]
    self.unified_stats = AnyStats()

  def add(self, l):
    self.total += 1
    if len(l) == 0:
      self.zero_len += 1
    else:
      self.min_len = min(self.min_len, len(l))
      self.max_len = max(self.max_len, len(l))
    for i, x in enumerate(l):
      if i < len(self.initial_stats):
        self.initial_stats[i].add(x)
      self.unified_stats.add(x)
    for i in range(i+1, len(self.initial_stats)):
      self.initial_stats[i].add_missing()

  def best_length(self):
    if self.zero_len == self.total:
      return 0
    elif self.min_len == self.max_len:
      return self.max_len
    elif self.min_len < len(self.initial_stats):
      return min(self.max_len, len(self.initial_stats))
    else:
      return self.min_len

  def paths(self, *, prefix=(), props=default_pathprops):
    # TODO: Should we do something about variable length?
    length = self.best_length()

    res = []
    for i in range(length):
      if i < len(self.initial_stats):
        res.append( self.initial_stats[i].paths(prefix=prefix+(i,), props=props) )
      else:
        res.append( self.unified_stats[i].paths(prefix=prefix+(i,), props=props) )

    whole_res = []
    for combo in product(*res):
      d = combo[0]
      for i in range(1, len(combo)):
        d.update(combo[i])
      whole_res.append(d)
    return whole_res

  def from_pathvals(self, pathvals):
    res = {}
    for k, v in pathvals.items():
      assert len(k) >= 1, "seq key too short"
      assert isinstance(k[0], int), "seq key not int"
      res.setdefault(k[0], {})[k[1:]] = v

    res2 = []
    for i in range(len(res)):
      assert i in res, "seq keys are non-contiguous"
      if i < len(self.initial_stats):
        res2.append(self.initial_stats[i].from_pathvals(res[i]))
      else:
        res2.append(self.unified_stats.from_pathvals(res[i]))
    return res2
