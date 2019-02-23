"""
Used by the DictStats object; may also be useful on their own
for solving the problem of valid parameter groupings.
"""

def find_cliques(G):  # Not currently using this, but keeping it around.
    """Returns all maximal cliques. Stolen from NetworkX, modified for directed graphs."""
    if len(G) == 0:
        return

    adj = {u: {v for v in G[u] if v != u and u in G[v]} for u in G}
    Q = [None]

    subg = set(G)
    cand = set(G)
    u = max(subg, key=lambda u: len(cand & adj[u]))
    ext_u = cand - adj[u]
    stack = []

    try:
        while True:
            if ext_u:
                q = ext_u.pop()
                cand.remove(q)
                Q[-1] = q
                adj_q = adj[q]
                subg_q = subg & adj_q
                if not subg_q:
                    yield Q[:]
                else:
                    cand_q = cand & adj_q
                    if cand_q:
                        stack.append((subg, cand, ext_u))
                        Q.append(None)
                        subg = subg_q
                        cand = cand_q
                        u = max(subg, key=lambda u: len(cand & adj[u]))
                        ext_u = cand - adj[u]
            else:
                Q.pop()
                subg, cand, ext_u = stack.pop()
    except IndexError:
        pass


def guess_key_groups_mo(sibling_keys, spousal_keys, alone_keys):
  """
  A heuristic for representing allowed groups of keys.

  It assumes that each group has a set of mandatory keys,
  and a set of optional keys. (This is not necessarily true,
  and this is not the only assumption this heuristic makes).

  Args:
    sibling_keys: dict. for each key (param name), other keys that sometimes appear with it.
    spousal_keys: dict. for each key (param name), other keys that it never appears without.
    alone_keys: set. keys that appear all by themselves.
  Return:
    iterator, yielding pairs of mandatory_keys and optional_keys sets for
    eacho prospective grouping.
  """
  # Start by assuming each key is the "root" of a mandatory group.
  groups = list(set(frozenset(i) for i in spousal_keys.values()))

  def compatible_optionals(group):
    return frozenset( k for k, v in sibling_keys.items()
                        if k not in group and group.issubset(v) )

  optionals = [compatible_optionals(i) for i in groups]
  unions = [(x|y) for x, y in zip(groups, optionals)]

  def should_skip(i):
    # Be very careful, it's really hard to add additional
    # skip conditions without breaking a case that "should" work.
    def other_groups():
      for j in range(len(groups)):
        if j == i:
          continue
        yield j

    if len(groups[i]) == 1:
      # Reject single-item groups that appear as optionals and
      # haven't appeared on their own. This is because it's
      # probably just a common optional.
      item = list(groups[i])[0]
      if any(item in optionals[j] for j in other_groups()):
        if item not in alone_keys:
          return True

    for j in other_groups():
      if unions[i] == unions[j] and groups[i] > groups[j]:
        # Same set of objects described, but by the other account
        # some stuff is optional. Prefer that one.
        return True
    return False

  # Filter out any groups that look like they were based around
  # something that should be an optional.
  for i in range(len(groups)):
    if not should_skip(i):
      yield groups[i], optionals[i]
