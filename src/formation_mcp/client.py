"""HTTP client for Formation REST API."""

from datetime import datetime, timedelta
from typing import Any

import httpx


class FormationClient:
    """Client for interacting with the Formation REST API.

    Supports two authentication modes:
    1. Token-based: Provide token directly
    2. Credential-based: Provide username/password, token obtained automatically
    """

    def __init__(
        self,
        base_url: str,
        token: str | None = None,
        username: str | None = None,
        password: str | None = None,
        timeout: float = 30.0,
    ):
        """Initialize the Formation API client.

        Args:
            base_url: Base URL of the Formation service
            token: JWT authentication token (optional if username/password provided)
            username: Username for automatic login (optional if token provided)
            password: Password for automatic login (optional if token provided)
            timeout: HTTP request timeout in seconds

        Raises:
            ValueError: If neither token nor username/password is provided
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self._token = token
        self._username = username
        self._password = password
        self._token_expiry: datetime | None = None

        if not token and not (username and password):
            raise ValueError("Must provide either token or username/password")

    async def login(self, username: str, password: str) -> dict[str, Any]:
        """Authenticate and obtain JWT token via HTTP Basic Auth.

        Args:
            username: User's username
            password: User's password

        Returns:
            Dictionary with access_token, token_type, expires_in

        Raises:
            httpx.HTTPStatusError: If authentication fails
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/login",
                auth=(username, password),
            )
            response.raise_for_status()
            return response.json()

    async def _ensure_token(self) -> None:
        """Ensure we have a valid token, logging in if necessary.

        Raises:
            ValueError: If no token and no credentials available
            httpx.HTTPStatusError: If login fails
        """
        # If we have a token and it's not expired, we're good
        if self._token:
            if not self._token_expiry or datetime.now() < self._token_expiry:
                return

        # Need to get a new token
        if self._username and self._password:
            result = await self.login(self._username, self._password)
            self._token = result["access_token"]
            # Set expiry to 60 seconds before actual expiry for safety
            expires_in = result.get("expires_in", 3600)
            self._token_expiry = datetime.now() + timedelta(seconds=expires_in - 60)
        else:
            raise ValueError("Token expired and no credentials available for refresh")

    @property
    def headers(self) -> dict[str, str]:
        """Get current authorization headers.

        Returns:
            Dictionary with Authorization header
        """
        if not self._token:
            return {}
        return {"Authorization": f"Bearer {self._token}"}

    async def list_apps(
        self,
        limit: int = 100,
        offset: int = 0,
        name: str | None = None,
    ) -> dict[str, Any]:
        """List interactive VICE applications.

        Args:
            limit: Maximum number of apps to return
            offset: Number of apps to skip for pagination
            name: Optional filter by app name

        Returns:
            Dictionary with 'total' and 'apps' list
        """
        await self._ensure_token()

        params: dict[str, Any] = {"limit": limit, "offset": offset}
        if name:
            params["name"] = name

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.get(
                f"{self.base_url}/apps",
                headers=self.headers,
                params=params,
            )
            response.raise_for_status()
            return response.json()

    async def get_app_parameters(self, app_id: str, system_id: str = "de") -> dict[str, Any]:
        """Get the parameters for an app.

        Args:
            app_id: UUID of the app
            system_id: System identifier (default: 'de')

        Returns:
            Dictionary with app parameter groups including required parameters
        """
        await self._ensure_token()

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.get(
                f"{self.base_url}/apps/{system_id}/{app_id}/parameters",
                headers=self.headers,
            )
            response.raise_for_status()
            return response.json()

    async def launch_app(
        self,
        app_id: str,
        system_id: str = "de",
        name: str | None = None,
        config: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Launch an interactive application.

        Args:
            app_id: UUID of the app to launch
            system_id: System identifier (default: 'de')
            name: Optional custom name for the analysis
            config: Configuration parameters for the app

        Returns:
            Dictionary with analysis_id, name, status, and optionally url
        """
        await self._ensure_token()

        body: dict[str, Any] = {}
        if name:
            body["name"] = name
        if config:
            body["config"] = config

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/app/launch/{system_id}/{app_id}",
                headers=self.headers,
                json=body if body else None,
            )
            response.raise_for_status()
            return response.json()

    async def get_analysis_status(self, analysis_id: str) -> dict[str, Any]:
        """Get the status of an analysis.

        Args:
            analysis_id: UUID of the analysis

        Returns:
            Dictionary with analysis_id, status, url_ready, and optionally url
        """
        await self._ensure_token()

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.get(
                f"{self.base_url}/apps/analyses/{analysis_id}/status",
                headers=self.headers,
            )
            response.raise_for_status()
            return response.json()

    async def list_analyses(self, status: str = "Running") -> dict[str, Any]:
        """List analyses filtered by status.

        Args:
            status: Status filter (default: 'Running'). Common values:
                   - 'Running': Currently executing
                   - 'Completed': Successfully finished
                   - 'Failed': Failed analyses
                   - 'Submitted': Queued but not running yet

        Returns:
            Dictionary with 'analyses' list
        """
        await self._ensure_token()

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.get(
                f"{self.base_url}/apps/analyses/",
                headers=self.headers,
                params={"status": status},
            )
            response.raise_for_status()
            return response.json()

    async def control_analysis(
        self, analysis_id: str, operation: str
    ) -> dict[str, Any]:
        """Control an analysis lifecycle.

        Args:
            analysis_id: UUID of the analysis
            operation: Control operation - one of:
                      - 'extend_time': Extend time limit
                      - 'save_and_exit': Save outputs and terminate
                      - 'exit': Terminate without saving

        Returns:
            Dictionary with operation result
        """
        await self._ensure_token()

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/apps/analyses/{analysis_id}/control",
                headers=self.headers,
                params={"operation": operation},
            )
            response.raise_for_status()
            return response.json()

    async def browse_data(
        self,
        path: str,
        offset: int = 0,
        limit: int | None = None,
        include_metadata: bool = False,
        avu_delimiter: str = ",",
    ) -> dict[str, Any]:
        """Browse iRODS directory or read file contents.

        Args:
            path: iRODS path to browse (can be directory or file)
            offset: Byte offset for file reading (default: 0)
            limit: Max bytes to read for files (default: None/all)
            include_metadata: Include iRODS AVU metadata in response (default: False)
            avu_delimiter: Delimiter for AVU metadata (default: ',')

        Returns:
            Dictionary with:
            - For directories: path, type, contents list
            - For files: content (bytes), headers (if metadata requested)
        """
        await self._ensure_token()

        params: dict[str, Any] = {}
        if offset:
            params["offset"] = offset
        if limit is not None:
            params["limit"] = limit
        if include_metadata:
            params["include_metadata"] = "true"
        if avu_delimiter != ",":
            params["avu_delimiter"] = avu_delimiter

        # Remove leading slash if present for the URL path
        url_path = path.lstrip("/")

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.get(
                f"{self.base_url}/data/{url_path}",
                headers=self.headers,
                params=params,
            )
            response.raise_for_status()

            # Check content type to determine if it's a directory listing or file
            content_type = response.headers.get("content-type", "")

            if "application/json" in content_type:
                # Directory listing
                return response.json()
            else:
                # File content - return as dict with content and metadata headers
                metadata_headers = {
                    k: v
                    for k, v in response.headers.items()
                    if k.lower().startswith("x-datastore-")
                }
                return {"content": response.content, "headers": metadata_headers}

    async def put_data(
        self,
        path: str,
        content: bytes | None = None,
        resource_type: str | None = None,
        metadata: dict[str, str] | None = None,
        replace_metadata: bool = False,
        avu_delimiter: str = ",",
    ) -> dict[str, Any]:
        """Create directory, upload file, or set metadata on iRODS path.

        Args:
            path: iRODS path (directory or file)
            content: File content as bytes (optional)
            resource_type: "directory" to create a directory without content
            metadata: Dict mapping attribute names to values (or "value,units" strings)
            replace_metadata: If True, replace existing metadata instead of adding
            avu_delimiter: Delimiter for separating value and units (default: ',')

        Returns:
            Dictionary with path, type, and created status
        """
        await self._ensure_token()

        params: dict[str, Any] = {}
        if resource_type:
            params["resource_type"] = resource_type
        if replace_metadata:
            params["replace_metadata"] = "true"
        if avu_delimiter != ",":
            params["avu_delimiter"] = avu_delimiter

        # Build headers with metadata
        headers = dict(self.headers)  # Start with auth headers
        if metadata:
            for attribute, value in metadata.items():
                headers[f"X-Datastore-{attribute}"] = value

        # Remove leading slash if present for the URL path
        url_path = path.lstrip("/")

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.put(
                f"{self.base_url}/data/{url_path}",
                headers=headers,
                params=params,
                content=content if content else b"",
            )
            response.raise_for_status()
            return response.json()

    async def delete_data(
        self,
        path: str,
        recurse: bool = False,
        dry_run: bool = False,
    ) -> dict[str, Any]:
        """Delete a file or directory from iRODS.

        Args:
            path: iRODS path to delete
            recurse: Allow deleting non-empty directories (default: False)
            dry_run: Preview deletion without executing (default: False)

        Returns:
            Dictionary with deletion result including dry_run status
        """
        await self._ensure_token()

        params: dict[str, Any] = {}
        if recurse:
            params["recurse"] = "true"
        if dry_run:
            params["dry_run"] = "true"

        # Remove leading slash if present for the URL path
        url_path = path.lstrip("/")

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.delete(
                f"{self.base_url}/data/{url_path}",
                headers=self.headers,
                params=params,
            )
            response.raise_for_status()
            return response.json()
