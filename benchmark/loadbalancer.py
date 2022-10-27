from typing import List, Callable
import math
import random

class Node:

    def __init__(self, tick_speed) -> None:
        self.requests: List[int] = []
        self.tick_speed = tick_speed

    def tick(self):
        tick_speed = self.tick_speed(self)
        if tick_speed < 0.01:
            return
        new_requests: List[int] = []
        for request in self.requests:
            if request > 0:
                new_requests.append(request - tick_speed)

        self.requests = new_requests

    def get_tick_speed(self):
        return self.tick_speed(self)

    def append(self, request: int):
        self.requests.append(request)


class Balancer:

    def __init__(self, balancing_algo: Callable[[str, List[Node]], int], nodes: List[Node]):
        self.balancing_algo = balancing_algo
        self.nodes = nodes

        # statistics
        self.max_requests_in_flight_all = 0
        self.max_requests_in_flight_single_node = 0
        self.min_tick_speed = 1


    def tick(self):
        for node in self.nodes:
            node.tick()

    def balance(self, request: int):
        node = self.balancing_algo(self.nodes)
        node.append(request)

        # collect stats
        self.max_requests_in_flight_all = max(self.max_requests_in_flight_all, sum(map(lambda n: len(n.requests), self.nodes)))
        self.max_requests_in_flight_single_node = max(self.max_requests_in_flight_single_node, len(node.requests))
        self.min_tick_speed = min(self.min_tick_speed, *map(lambda node: node.get_tick_speed(), self.nodes))

def make_tick_speed(max_requests_before_disaster: int):
    def tick_speed(self):
        # HACK short circuit to avoid huge numbers on math.exp(), which breaks when near 750.
        # We could instead use math.e ** x instead of math.exp(x), but that's considerably slower.
        if len(self.requests) >= 500:
            return 0
        return round(1 / (1 + math.exp(len(self.requests) - max_requests_before_disaster)), 2)

    return tick_speed

### Algorithm section

def random_choice(nodes: List[Node]) -> Node:
    return random.choice(nodes)

def two_choices(nodes: List[Node]) -> Node:
    a = random.choice(nodes)
    b = random.choice(nodes)

    return a if len(a.requests) <= len(b.requests) else b

def make_round_robin():
    idx = 0
    def round_robin(nodes: List[Node]) -> Node:
        nonlocal idx
        idx = (idx + 1) % len(nodes)
        return nodes[idx]

    return round_robin

all_algorithms = [random_choice, two_choices, make_round_robin()]

### End Algorithm section

if __name__ == '__main__':
    for load_algorithm in all_algorithms:
        print(load_algorithm.__name__)
        tick_speed = make_tick_speed(max_requests_before_disaster = 10)

        nodes = [Node(tick_speed) for _ in range(100)]
        balancer = Balancer(load_algorithm, nodes)

        for request_idx in range(10000):
            request_size = random.randint(100, 200)
            balancer.balance(request_size)
            balancer.tick()

        print(f'\tmax_requests = {balancer.max_requests_in_flight_all} max_requests_in_flight_single_node = {balancer.max_requests_in_flight_single_node} min_tick_speed = {balancer.min_tick_speed}')
        print('\n')