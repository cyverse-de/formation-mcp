"""High-level workflows combining multiple Formation API operations."""

import asyncio
import time
import webbrowser
from typing import Any

from .client import FormationClient


class FormationWorkflows:
    """High-level workflows for Formation operations."""

    def __init__(
        self,
        base_url: str,
        token: str | None = None,
        username: str | None = None,
        password: str | None = None,
    ):
        """Initialize workflows with a Formation API client.

        Args:
            base_url: Base URL of the Formation service
            token: JWT authentication token (optional if username/password provided)
            username: Username for automatic login (optional if token provided)
            password: Password for automatic login (optional if token provided)
        """
        self.client = FormationClient(
            base_url=base_url,
            token=token,
            username=username,
            password=password,
        )

    async def launch_and_wait(
        self,
        app_id: str,
        system_id: str = "de",
        name: str | None = None,
        max_wait: int = 300,
        poll_interval: int = 5,
        config: dict[str, Any] | None = None,
        overall_job_type: str | None = None,
    ) -> dict[str, Any]:
        """Launch an app and wait for it to become ready.

        Args:
            app_id: UUID of the app to launch
            system_id: System identifier (default: 'de')
            name: Optional custom name for the analysis
            max_wait: Maximum seconds to wait for readiness (default: 300)
            poll_interval: Seconds between status polls (default: 5)
            config: Launch configuration including required parameters
            overall_job_type: Job type from list_apps (optional).
                             If provided, skips parameter fetch.
                             Can be: "Interactive" (VICE), "DE" (batch), "OSG", or "Tapis"

        Returns:
            Dictionary with analysis_id, url, status, and wait_time (for interactive apps), or
            Dictionary with analysis_id and status (for batch jobs), or
            Dictionary with missing_params and config_template if params needed

        Raises:
            TimeoutError: If interactive app doesn't become ready within max_wait seconds
            httpx.HTTPStatusError: If API requests fail
        """
        # If overall_job_type is provided, skip parameter fetch and validation
        # Otherwise, get app configuration to check for required parameters and job type
        if overall_job_type is None:
            app_config = await self.client.get_app_parameters(app_id, system_id)

            # Check if there are required parameters
            missing_params = self._check_required_params(app_config, config or {})

            if missing_params:
                return {
                    "missing_params": missing_params,
                    "config_template": app_config,
                }

            job_type = app_config.get("overall_job_type", "")
        else:
            job_type = overall_job_type

        # Determine if this is an interactive/VICE app
        is_interactive = job_type == "Interactive"

        # Launch the app with the provided config
        launch_result = await self.client.launch_app(
            app_id=app_id,
            system_id=system_id,
            name=name,
            config=config,
        )
        analysis_id = launch_result["analysis_id"]

        # For batch jobs, return immediately without waiting for URL
        if not is_interactive:
            return {
                "analysis_id": analysis_id,
                "status": "submitted",
                "job_type": job_type,
            }

        # For interactive apps, poll status until ready or timeout
        start_time = time.time()
        while time.time() - start_time < max_wait:
            status_result = await self.client.get_analysis_status(analysis_id)

            if status_result.get("url_ready"):
                wait_time = int(time.time() - start_time)
                return {
                    "analysis_id": analysis_id,
                    "url": status_result["url"],
                    "status": status_result.get("status", "ready"),
                    "wait_time": wait_time,
                }

            # Wait before polling again
            await asyncio.sleep(poll_interval)

        # Timeout - return status but indicate not ready
        elapsed = int(time.time() - start_time)
        raise TimeoutError(
            f"Analysis {analysis_id} not ready after {elapsed}s (max: {max_wait}s)"
        )

    def open_in_browser(self, url: str) -> dict[str, Any]:
        """Open a URL in the default web browser.

        Args:
            url: URL to open

        Returns:
            Dictionary with success status and url
        """
        try:
            webbrowser.open(url)
            return {"success": True, "url": url}
        except Exception as e:
            return {"success": False, "url": url, "error": str(e)}

    async def get_running_analyses(self) -> list[dict[str, Any]]:
        """Get list of all running analyses.

        Returns:
            List of running analyses with analysis_id, app_id, system_id, status
        """
        result = await self.client.list_analyses(status="Running")
        return result.get("analyses", [])

    async def stop_analysis(
        self, analysis_id: str, save_outputs: bool = True
    ) -> dict[str, Any]:
        """Stop a running analysis.

        Args:
            analysis_id: UUID of the analysis to stop
            save_outputs: Whether to save outputs before stopping (default: True)

        Returns:
            Dictionary with termination result
        """
        operation = "save_and_exit" if save_outputs else "exit"
        return await self.client.control_analysis(analysis_id, operation)

    async def extend_analysis_time(self, analysis_id: str) -> dict[str, Any]:
        """Extend the time limit for a running analysis.

        Args:
            analysis_id: UUID of the analysis

        Returns:
            Dictionary with new time limit information
        """
        return await self.client.control_analysis(analysis_id, "extend_time")

    def _check_required_params(
        self, app_config: dict[str, Any], provided_config: dict[str, Any]
    ) -> list[dict[str, Any]]:
        """Check if all required parameters are provided.

        Args:
            app_config: App configuration from the API
            provided_config: User-provided configuration

        Returns:
            List of missing required parameters with their details
        """
        missing = []
        groups = app_config.get("groups", [])

        for group in groups:
            for param in group.get("parameters", []):
                param_id = param.get("id")
                is_required = param.get("isVisible", True) is not False and param.get(
                    "required", False
                )

                if is_required and param_id not in provided_config:
                    missing.append(
                        {
                            "id": param_id,
                            "name": param.get("name", param_id),
                            "description": param.get("description", ""),
                            "type": param.get("type", "string"),
                        }
                    )

        return missing
