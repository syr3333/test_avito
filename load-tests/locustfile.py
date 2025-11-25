import random
import string
import threading
from typing import List, Tuple

import requests
from locust import HttpUser, task, between, events


TEAM_MEMBER_COUNT = 10
TEAM_POOL_SIZE = 20
WAIT_MIN_SECONDS = 0.05
WAIT_MAX_SECONDS = 0.3
DEFAULT_HOST = "http://localhost:8080"

TeamRecord = Tuple[str, Tuple[str, ...]]

_TEAM_POOL: List[TeamRecord] = []
_TEAM_LOCK = threading.Lock()


def _random_id(length: int = 8) -> str:
    alphabet = string.ascii_lowercase + string.digits
    return ''.join(random.choices(alphabet, k=length))


# Pre-create a bounded pool of teams so high-concurrency runs do not fight over team creation.
def _bootstrap_team_pool(host: str) -> None:
    base_url = host or DEFAULT_HOST

    with _TEAM_LOCK:
        if _TEAM_POOL:
            return

        session = requests.Session()
        for _ in range(TEAM_POOL_SIZE):
            suffix = _random_id()
            team_name = f"team-{suffix}"
            members = tuple(f"user-{team_name}-{i}" for i in range(TEAM_MEMBER_COUNT))
            payload = {
                "team_name": team_name,
                "members": [
                    {"user_id": user_id, "username": f"User {idx}", "is_active": True}
                    for idx, user_id in enumerate(members)
                ],
            }

            resp = session.post(f"{base_url}/team/add", json=payload, timeout=10)
            if resp.status_code != 201:
                raise RuntimeError(
                    f"Failed to create team {team_name}: {resp.status_code} {resp.text}"
                )

            _TEAM_POOL.append((team_name, members))


def _pick_team() -> TeamRecord:
    if not _TEAM_POOL:
        raise RuntimeError("Team pool not initialized; ensure test_start ran")
    return random.choice(_TEAM_POOL)


class PRServiceUser(HttpUser):
    """Synthetic client that mirrors the way product teams use the service."""

    wait_time = between(WAIT_MIN_SECONDS, WAIT_MAX_SECONDS)

    def on_start(self) -> None:
        team_name, members = _pick_team()
        self.team_name = team_name
        self.users = list(members)

    def _pick_user(self) -> str | None:
        return random.choice(self.users) if self.users else None

    @task(10)
    def fetch_statistics(self) -> None:
        self.client.get("/statistics")

    @task(5)
    def fetch_team(self) -> None:
        self.client.get(f"/team/get?team_name={self.team_name}")

    @task(8)
    def create_pull_request(self) -> None:
        author = self._pick_user()
        if not author:
            return

        pr_id = f"pr-{_random_id()}"
        self.client.post(
            "/pullRequest/create",
            json={
                "pull_request_id": pr_id,
                "pull_request_name": f"Feature {pr_id}",
                "author_id": author,
            },
        )

    @task(3)
    def fetch_user_reviews(self) -> None:
        reviewer = self._pick_user()
        if not reviewer:
            return

        self.client.get(f"/users/getReview?user_id={reviewer}")

    @task(2)
    def merge_pull_request(self) -> None:
        author = self._pick_user()
        if not author:
            return

        pr_id = f"pr-merge-{_random_id()}"
        create_response = self.client.post(
            "/pullRequest/create",
            json={
                "pull_request_id": pr_id,
                "pull_request_name": f"Hotfix {pr_id}",
                "author_id": author,
            },
            name="/pullRequest/create (merge)",
        )

        if create_response.status_code == 201:
            self.client.post(
                "/pullRequest/merge",
                json={"pull_request_id": pr_id},
            )

    @task(1)
    def toggle_user_activity(self) -> None:
        user_id = self._pick_user()
        if not user_id:
            return

        self.client.post(
            "/users/setIsActive",
            json={
                "user_id": user_id,
                "is_active": random.choice([True, False]),
            },
        )


@events.test_start.add_listener
def on_test_start(environment, **_) -> None:
    print("Starting load tests")
    print(f"Target: {environment.host}")
    _bootstrap_team_pool(environment.host)


@events.test_stop.add_listener
def on_test_stop(environment, **_) -> None:
    stats = environment.stats.total
    print("\nTests completed")
    print(f"Requests: {stats.num_requests}")
    print(f"Failures: {stats.num_failures}")
    if stats.num_requests:
        success_rate = (1 - stats.num_failures / stats.num_requests) * 100
        print(f"Success rate: {success_rate:.2f}%")
