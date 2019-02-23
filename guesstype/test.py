import pprint
import random
from .stats import *
from . import easy

def _test_dict_stats_mo_helper(dicts, expect):
  x = DictStats()
  for i in dicts:
    x.add(i)
  actual = set(x.guess_key_groups_mo())
  print('----------')
  for m, o in actual:
    print('\t', set(m), '\t', set(o))
  for i in expect:
    assert i in actual, "Missing %r" % (i,)

def test_dict_stats_mo():
  # For debugging guess_key_groups_mo, it's hard.
  _test_dict_stats_mo_helper([
      {'objAx': 1, 'objAy': 2, 'objAoptx': 3, 'objAopty': 4, 'common': 5},
      {'objAx': 1, 'objAy': 2, 'common': 3},
      {'objBx': 1, 'objBy': 2, 'common': 3},
      {'objBx': 1, 'objBy': 2, 'objBoptx': 3, 'objBopty': 3},
      {'objBx': 1, 'objBy': 2},
    ], [
      (frozenset({'objAx', 'objAy', 'common'}), frozenset({'objAoptx', 'objAopty'})),
      (frozenset({'objBx', 'objBy'}), frozenset({'common', 'objBoptx', 'objBopty'})),
  ])

  _test_dict_stats_mo_helper([
     {'metha': 1, 'arg1': 2, 'arg2': 3, 'arg3': 4, 'arg4': 5},
     {'methb': 1, 'arg1': 2, 'arg2': 3},
     {'methb': 1, 'arg1': 2},
    ], [
      (frozenset({'methb', 'arg1'}), frozenset({'arg2'})),
  ])

  _test_dict_stats_mo_helper([
      {'always': 1, 'a': 2, 'b': 3},
      {'always': 1, 'b': 2},
      {'always': 1, 'c': 2, 'd': 3},
    ], [
      (frozenset({'always'}), frozenset({'a', 'b', 'c', 'd'})),
  ])

  _test_dict_stats_mo_helper([
     {'something': 1, 'arg1': 2, 'arg2': 3, 'arg3': 4, 'arg4': 5},
     {'something': 1, 'arg1': 2, 'arg2': 3},
    ], [
      (frozenset({'something', 'arg1', 'arg2'}), frozenset({'arg3', 'arg4'})),
  ])

def test_stuff():
  # Round trip some crap.
  S = AnyStats()
  S.add('{"k":"420blazeit", "k2": "<20, 21, 22, 23>"}')
  S.add('{"k":"420blazeit"}')
  path = S.paths(props=PathProperties(mono=True))
  assert len(path) == 1
  path = path[0]
  test = {}
  R = random.Random()
  for k, v in path.items():
    print(k, v)
    test[k] = v.gen_ident(R)
  print(S.from_pathvals(test))

def test_all():
  test_dict_stats_mo()
  test_stuff()
  pprint.pprint(easy.decode_one('{"numbers":"867 4678 23 34", "b64":"aGVsbG8K"}'))
