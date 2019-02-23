import json
from collections import namedtuple, Counter
import string, base64, binascii, zlib
import datetime

#
# Helpers
#

class cached_property(property):
  def __get__(self, instance, owner):
    if instance is None:
      return super().__get__(instance, owner)

    s = '_cached_' + self.fget.__name__
    try: return getattr(instance, s)
    except AttributeError: pass

    ret = super().__get__(instance, owner)
    setattr(instance, s, ret)
    return ret

  def __set__(self, instance, value):
    raise AttributeError()

  def __delete__(self, instance):
    if instance is not None:
      s = '_cached_' + self.fget.__name__
      try: delattr(instance, s)
      except AttributeError: pass
    super().__delete__(instance)


class count_trues(object):
  """Kinda like sum(), but only for the specific case of taking a sum()
  of booleans and then applying comparison operators. The advantage is
  that it doesn't go through the whole iterator if it doesn't need to.
  For example, in `t = count_trues(i%2 for i in range(99999))`, comparing
  `t` with 8 using any comparison operator will take 16 to 18 iterations.
  """
  def __init__(self, iterable):
    self.count = 0
    self.iterable = iterable

  def __gt__(self, other):
    if self.count > other:
      return True
    for i in self.iterable:
      if i:
        self.count += 1
      if self.count > other:
        return True
    return False

  def __ge__(self, other):
    if self.count >= other:
      return True
    for i in self.iterable:
      if i:
        self.count += 1
      if self.count >= other:
        return True
    return False

  def __lt__(self, other):
    return not self.__ge__(other)

  def __le__(self, other):
    return not self.__gt__(other)

  def __eq__(self, other):
    return self.__ge__(other) and self.__le__(other)

  def __ne__(self, other):
    return not self.__eq__(other)


def maybe_decode_base64(s):
  """Returns None if s doesn't look like base64 (even if it's decodeable).
  Otherwise, return the decoded string and the altchars used."""
  s = s.strip()
  charset = frozenset(s)
  if len(s) < 4:
    return None
  if not charset.issubset(string.ascii_letters + string.digits + '+/-_.=\n'):
    return None
  if len(charset.intersection('+/-_.=')) > 3:
    return None

  # Looks decodable, lets decode it.
  altchars = ''.join(charset.intersection('+/_-.'))[:2]
  altchars = sorted(altchars)
  if len(altchars) == 0:
    altchars = '+/'
  elif len(altchars) == 1:
    altchars = {
      '+': '+/',
      '/': '+/',
      '-': '-_',
      '_': '-_',
      '.': '._',
    }[altchars]

  try:
    decoded = base64.b64decode(s + '===', altchars)
  except:
    return None

  # Ending in the right number of '='s is a good tell.
  if s.endswith('=') and not s.endswith('===') and ('=' not in s.rstrip('=')):
    return decoded, altchars

  # So is decoding to all printables or mostly letters.
  if len(decoded) > 10 and charset.issubset(string.printable):
    return decoded, altchars
  if len(decoded) >= 4:
    good_chars = frozenset(string.ascii_letters + string.digits + ' _.,')
    if sum(c in good_chars for c in decoded) > len(decoded) * 0.8:
      return decoded, altchars

  # Super good compressability is also a sign.
  if float(len(zlib.compress(decoded))) / len(decoded) <= 0.40:
    return decoded, altchars

  # decoded is apparently binary junk. Make sure the input really
  # "looks like" base64 before returning it. Specifically, we
  # check that the breakdown of digits to lowercase to capitals looks
  # roughly even.
  if not charset.intersection(string.digits): return None
  if not charset.intersection(string.ascii_lowercase): return None
  if not charset.intersection(string.ascii_uppercase): return None

  count = Counter(s)
  del count['\n']
  del count['=']
  del count['A'] # Strings of null bytes are often more common, that's fine.
  grand_total = sum(count.values())
  digits = sum(count[c] for c in string.digits)
  lower = sum(count[c] for c in string.ascii_lowercase)
  upper = sum(count[c] for c in string.ascii_uppercase)
  total = digits + lower + upper
  if total < grand_total * 0.9: # Too many other chars to be base64!
    return None

  # Multipled by inverse expected probability of seeing that letter.
  digits = digits / 10.0 / total * 63
  lower = lower / 26.0 / total * 63
  upper = upper / 25.0 / total * 63
  if all(i > 0.5 for i in (digits, lower, upper)):
    return decoded, altchars


#
# Classes for representing the "inferred type" of a single string.
#

class StrTypeBase:
  __slots__ = ()

class StrBool(namedtuple('StrBool', 'true false'), StrTypeBase):
  def invert(self, val):
    return self.true if val else self.false
  __slots__ = ()

class StrDate(namedtuple('StrDate', 'kind'), StrTypeBase):
  def invert(self, val):
    return val.isoformat()
  __slots__ = ()

class StrJSON(namedtuple('StrJSON', 'type'), StrTypeBase):
  def invert(self, val):
    return json.dumps(val)
  __slots__ = ()

class StrHex(namedtuple('StrHex', 'upper'), StrTypeBase):
  def invert(self, val):
    if isinstance(val, str):
      val = val.encode('utf-8')
    return binascii.hexlify(val).decode('utf-8')
  __slots__ = ()

class StrB64(namedtuple('StrB64', 'altchars'), StrTypeBase):
  def invert(self, val):
    # Lots of utf-8 hax. We don't expect to handle bytes well anyways.
    if isinstance(val, str):
      val = val.encode('utf-8')
    return base64.b64encode(val, altchars=self.altchars.encode('utf-8')).decode('utf-8')
  __slots__ = ()

class StrNum(namedtuple('StrNum', ''), StrTypeBase):
  def invert(self, val):
    return str(val)
  __slots__ = ()

class StrList(namedtuple('StrList', 'before delimiter after'), StrTypeBase):
  def invert(self, val):
    return self.before + self.delimiter.join(val) + self.after
  __slots__ = ()

class StrStr(namedtuple('StrStr', ''), StrTypeBase):
  def invert(self, val):
    return val
  __slots__ = ()

class StringGuesser:
  """Class for staring really hard at strings and deciding what they are."""
  def __init__(self, s):
    self.s = s
    self.stripped = s.strip()
    self.stripped_lower = self.stripped.lower()

  # Currently, the properties are non-None if the value can be interpreted
  # as the given type and **is likely to be the given type**. (For example,
  # "abcd" will currently return None on .hexbytes).
  # This is kinda bad cuz interpretation should be separate from guessing,
  # but whatever, not going to change this right now.

  @cached_property
  def isodate(self):
    # Parse ISO dates, ignoring fractional seconds or timezones cuz I'm lazy.
    stripped = self.stripped

    if ':' not in stripped:
      return None, None

    # Find end before fractional seconds or timezone
    colon = stripped.index(':')
    end = len(stripped)
    for c in '.Z+- ':
      try:
        end = min(end, stripped.index(c, colon))
      except (IndexError, ValueError):
        continue
    # Sanity check that after-the-end looks close enough.
    # Probably gonna be just a colon left or timezone name.
    if len(stripped[end:].strip(' .0123456789Z+-')) <= 4:
      try:
        val = datetime.datetime.strptime(stripped[:end], "%Y-%m-%dT%H:%M:%S")
      except ValueError:
        return None, None
      return StrDate('iso'), val

    return None, None

  @cached_property
  def boolean(self):
    if self.stripped_lower not in ("true", "yes", "false", "no"):
      return None, None

    info = (
      ("true", "false"),
      ("True", "False"),
      ("TRUE", "FALSE"),
      ("yes", "no"),
      ("Yes", "No"),
      ("YES", "NO"),
    )
    for i in info:
      if self.stripped in i:
        return StrBool(*i), self.stripped == i[0]
    else:
      # weird, couldn't match case right, just use lowercase.
      if self.stripped_lower in i:
        return StrBool(*i), self.stripped_lower == i[0]
    assert False, "Unreachable"

  @cached_property
  def hexbytes(self):
    cset = frozenset(self.stripped_lower)
    if self.stripped_lower.count(' ') > len(self.stripped_lower)/8:
      return None, None
    if len(cset) >= 3 and cset.issubset('0123456789abcdef _-') and not cset.issubset('abcdef _-'):
      try:
        decoded = binascii.unhexlify(self.stripped_lower.replace(' ', '').replace('_','').replace('-',''))
      except:
        return None, None

      # StrHex doesn't represent any space/underscore/dash patterns
      upper = 0
      for c in self.stripped:
        if c in 'abcdef':
          upper -= 1
        elif c in 'ABCDEF':
          upper += 1
      return StrHex(upper > 0), decoded

    return None, None

  @cached_property
  def base64bytes(self):
    ret = maybe_decode_base64(self.stripped)
    if ret is None:
      return None, None
    decoded, altchars = ret
    return StrB64(altchars), decoded

  @cached_property
  def json(self):
    try:
      x = json.loads(self.s)
    except:
      return None, None
    else:
      typename = type(x).__name__
      # We handle non-compoud json types with the other things.
      if typename not in ('dict', 'list'):
        return None, None
      return StrJSON(type=typename), x

  @cached_property
  def number(self):
    asint, asfloat = None, None
    try:
      asint = int(self.stripped)
    except ValueError:
      pass
    try:
      asfloat = float(self.stripped)
    except ValueError:
      pass

    if asint is not None:
      return StrNum(), asint
    elif asfloat is not None:
      return StrNum(), asfloat
    else:
      return None, None

  @cached_property
  def list(self):
    def comma_list(s, delimiter):
      if delimiter not in s:
        return None
      before, after = '', ''
      if s[:1] in '([<' and s[-1:] in ')]>':
        before, after = s[:1], s[-1:]
        s = s[1:-1]

      res = [StringGuesser(i) for i in s.split(delimiter)]
      if len(res) < 4 and count_trues(i.nonbasic for i in res) < 2:
        # If there's not many elements and less than two look like anything,
        # it's probably just a string with the delimiter in it.
        return None
      return StrList(before, delimiter, after), res

    def newline_list(s):
      after = '\n' if s.endswith('\n') else ''
      s = s.strip('\n')
      if '\n' not in s:
        return None
      res = [StringGuesser(i) for i in s.split('\n')]
      if len(res) >= 4 and count_trues(i.basic for i in res) >= len(res)/2:
        # If over half the lines seem like just strings, it's
        # probably just a string with newlines in it.
        return None
      return StrList('', '\n', after), res

    def whitespace_list(s):
      after = '\n' if s.endswith('\n') else ''
      s = s.strip().split()
      if len(s) <= 1:
        return None
      res = [StringGuesser(i) for i in s]
      if count_trues(i.basic for i in res) >= len(res)/2:
        # If over half the lines seem like just strings, it's
        # probably just a string with spaces in it.
        return None
      # fixme: ...we just assume space was the delimiter. Could do better.
      return StrList('', ' ', after), res

    # Trying them serially like this seems to makes sense; we could
    # return multiple possibilities but it doesn't seem worth it.
    s = self.s
    return newline_list(s) or comma_list(s, '\t') or comma_list(s, ':') or comma_list(s, ',') or whitespace_list(s) or (None, None)

  @property
  def nonbasic(self):
    """True if we get any sort of inference result beyond 'just some string'."""
    return not self.basic

  @property
  def basic(self):
    """True if there's no inference result besides 'just some string'."""
    return isinstance(self.best_type, StrStr)

  @property
  def best(self):
    if self.boolean[0] is not None: return self.boolean
    if self.number[0] is not None: return self.number
    if self.isodate[0] is not None: return self.isodate
    if self.hexbytes[0] is not None: return self.hexbytes
    if self.base64bytes[0] is not None: return self.base64bytes
    if self.json[0] is not None: return self.json
    if self.list[0] is not None: return self.list
    return StrStr(), self.s

  @property
  def best_type(self):
    return self.best[0]
