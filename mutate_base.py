import random


class Mutator(object):
    def __init__(self):
        pass

    def collect_deltas(self, target, route):
        raise Exception("Not implemented")

    def got_new_cov(self):
        raise Exception("Not implemented")

    def mutate(self):
        raise Exception("Not implemented")

    def current_route(self):
        raise Exception("Not implemented")

    def next_route(self):
        raise Exception("Not implemented")

    def on_response(self, target, status_code):
        route = self.current_route()
        self.collect_deltas(target, route)
        # Unconditionally mutate so we have different params next time we hit
        # this route
        self.mutate()

        self._print_code_diff(status_code, route)

    def _print_code_diff(self, code, route):
        if code != route.har_status_code:
            print(
                "\n\n\tActual Status code: {} browser: {} har replay: {}\n\n".format(
                    code, route.browser_status_code, route.har_status_code
                )
            )


class InfiniteMutator(Mutator):
    def __init__(self, routes):
        self.routes = routes
        self.route_index = -1
        self.phase = 0

    def next_route(self, skip_current_route=False):
        # Continue mutating the current route
        if self.got_new_cov() and not skip_current_route:
            return self.current_route()

        # Pick the next route
        if self.phase == 1:
            self._randomize_route()
        else:
            self.route_index += 1

        # Only do this check if we are in finite fuzz phase
        if self.phase == 0 and len(self.routes) <= self.route_index:
            return None
            self.phase = 1
            self._randomize_route()

        return self.current_route()

    def _randomize_route(self):
        self.route_index = random.randint(0, len(self.routes) - 1)

    def current_route(self):
        return self.routes[self.route_index]


class FiniteMutator(Mutator):
    def __init__(self, routes):
        self.routes = routes
        self.route_index = -1

    def next_route(self, skip_current_route=False):
        # Continue mutating the current route
        if self.got_new_cov() and not skip_current_route:
            return self.current_route()

        self.route_index += 1
        if len(self.routes) <= self.route_index:
            return None
        return self.current_route()

    def current_route(self):
        return self.routes[self.route_index]
