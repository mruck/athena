import random

import fuzzer.mutate_base as mutate_base
import mutation_state
import fuzzer.params as params_lib
import query
import query_metadata as query_metadata_lib

HAR_REPLAY = 0
ALL = 1
RANDOM = 2


# Params are assumed to have been present in queries
def mutate_params_in_queries(params):
    # Get a list of failed queries for each param
    failed_queries = []
    for p in params:
        failed = [q for q in p.query_metadata_list[0] if not q.query_obj.successful]
        # This param has failed queries
        if len(failed) > 0:
            failed_queries.append((p, failed))

    # Queries failed with params we control.  Mutate those.
    if len(failed_queries) > 0:
        # Pick the param from the first failed query and mutate it
        print("\t***Failed queries***")
        for param, queries in failed_queries:
            print(
                "\tParam %s has %d failed queries" % (param.name.upper(), len(queries))
            )
            print(
                "\tselect %s from %s"
                % (queries[0].col.upper(), queries[0].table.upper())
            )
            print(query.stringify_query(queries[0].query_obj))
            # Update param.next_val using a value from the db
            val = param.update_next_val(table=queries[0].table, col=queries[0].col)
            print("\tnext val: {}\n".format(val))
    # No queries failed that we control.  Pick another valid value from the db.
    else:
        for param in params:
            print("Mutating %s" % param.name)
            # Grab the first query
            q = param.query_metadata_list[0][0]
            # Update param.next_val using a value from the db
            val = param.update_next_val(table=q.table, col=q.col)
            print("next val: {}".format(val))


class NaiveInfiniteMutator(mutate_base.InfiniteMutator):
    def __init__(
        self, har_routes, all_routes, stop_after_har=False, stop_after_all_routes=False
    ):
        self.har_routes = har_routes
        self.all_routes = all_routes
        self.route_index = -1
        self.src_delta = False
        self.params_delta = False
        self.queries_delta = False
        self.phase = HAR_REPLAY
        self.stop_after_har = stop_after_har
        self.stop_after_all_routes = stop_after_all_routes
        self.skip_current = False

    def collect_deltas(self, target, route):
        # Read queries and params dumped by rails and update the route obj
        mutation_state.update_route_state(target, route)
        self.src_delta = len(target.cov.update()) > 0
        self.params_delta = params_lib.params_delta(
            route.query_params + route.body_params + route.dynamic_segments
        )
        # TODO: Keep a list of unique queries in the route obj
        self.queries_delta = query.queries_delta(route.queries[0], route.unique_queries)

    def next_route(self):
        if self.phase == HAR_REPLAY:
            # Continue fuzzing the current route
            if self.got_new_cov() and not self.skip_current:
                return self.current_route()

            # Next route please
            self.skip_current = False
            self.route_index += 1
            if self.route_index < len(self.har_routes):
                return self.current_route()

            # We are done mutating har requests, should we stop?
            if self.stop_after_har:
                return None

            # We are in the next phase
            self.phase = ALL
            self.route_index = 0
            print("Entering ALL phase")
            return self.current_route()

        if self.phase == ALL:
            # Continue fuzzing the current route
            if self.got_new_cov() and not self.skip_current:
                return self.current_route()

            # Next route please
            self.skip_current = False
            self.route_index += 1
            if self.route_index < len(self.all_routes):
                return self.current_route()

            # We have hit all routes, should we stop?
            if self.stop_after_all_routes:
                return None

            # We exhausted all routes.  Pick at random
            self.phase = RANDOM
            print("Entering RANDOM phase")

        # We are in the random phase
        self._randomize_route()
        return self.current_route()

    def _randomize_route(self):
        self.route_index = random.randint(0, len(self.all_routes) - 1)

    def current_route(self):
        if self.phase == HAR_REPLAY:
            return self.har_routes[self.route_index]
        return self.all_routes[self.route_index]

    def got_new_cov(self):
        return self.src_delta or self.params_delta or self.queries_delta

    def mutate(self):
        # In place param mutations
        route = self.current_route()
        params = route.body_params + route.query_params + route.dynamic_segments

        # We got new cov.  Check if params are present in queries and if so mutate them
        if self.src_delta:
            # Check if most recent request made queries.
            if len(route.queries[0]) > 0:
                # Get a list of parameters that showed up in queries made by the most
                # recent request. This does an in place update for each param obj
                query_metadata_lib.search_queries_for_params(params, route.queries[0])
                # Return params present in most queries
                filtered_params = query_metadata_lib.check_for_new_queries(params)
                # Params were present in queries
                if len(filtered_params) > 0:
                    # Mutate params that showed up in queries
                    mutate_params_in_queries(filtered_params)
                    return
            # Params were not present in queries, default to naive mutation
            # and send original har values if present
            [p.mutate(respect_har=True) for p in params]
        # We have exhausted query mutation.  Try mutating params we discovered.
        else:
            [p.mutate() for p in params]

    def force_next_route(self):
        self.skip_current = True
