"""Configuration management for Formation MCP server."""

import os
from dataclasses import dataclass


@dataclass
class Config:
    """Configuration loaded from environment variables.

    Supports two authentication modes:
    1. Token-based: Provide FORMATION_TOKEN directly
    2. Credential-based: Provide FORMATION_USERNAME and FORMATION_PASSWORD
       (token will be obtained automatically)
    """

    base_url: str
    token: str | None = None
    username: str | None = None
    password: str | None = None

    @classmethod
    def from_env(cls) -> "Config":
        """Load configuration from environment variables.

        Requires FORMATION_BASE_URL plus either:
        - FORMATION_TOKEN (pre-obtained JWT token), or
        - FORMATION_USERNAME and FORMATION_PASSWORD (for automatic login)

        Raises:
            ValueError: If required environment variables are missing
        """
        base_url = os.environ.get("FORMATION_BASE_URL", "")
        token = os.environ.get("FORMATION_TOKEN")
        username = os.environ.get("FORMATION_USERNAME")
        password = os.environ.get("FORMATION_PASSWORD")

        if not base_url:
            raise ValueError("FORMATION_BASE_URL environment variable is required")

        # Must provide either token OR username+password
        has_token = bool(token)
        has_credentials = bool(username and password)

        if not has_token and not has_credentials:
            raise ValueError(
                "Must provide either FORMATION_TOKEN or both "
                "FORMATION_USERNAME and FORMATION_PASSWORD"
            )

        return cls(
            base_url=base_url.rstrip("/"),
            token=token,
            username=username,
            password=password,
        )


# Global config instance - loaded when module is imported
config = Config.from_env()
