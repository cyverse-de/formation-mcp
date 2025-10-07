"""Tests for Formation workflows."""

import pytest
from pytest_httpx import HTTPXMock

from formation_mcp.workflows import FormationWorkflows


@pytest.fixture
def workflows():
    """Create test workflows instance."""
    return FormationWorkflows(
        base_url="https://formation.test",
        token="test-token",
    )


@pytest.mark.asyncio
async def test_launch_and_wait_immediate_url(
    workflows: FormationWorkflows, httpx_mock: HTTPXMock
):
    """Test launch and wait when URL becomes ready immediately."""
    app_id = "app-123"
    analysis_id = "analysis-456"

    # Mock app config with no required parameters
    httpx_mock.add_response(
        url=f"https://formation.test/apps/de/{app_id}/config",
        json={"groups": []},
    )

    # Mock launch response
    httpx_mock.add_response(
        url=f"https://formation.test/app/launch/de/{app_id}",
        json={
            "analysis_id": analysis_id,
            "name": "Test Analysis",
            "status": "Running",
        },
        method="POST",
    )

    # Mock status check - URL ready immediately
    httpx_mock.add_response(
        url=f"https://formation.test/apps/analyses/{analysis_id}/status",
        json={
            "analysis_id": analysis_id,
            "status": "Running",
            "url_ready": True,
            "url": "https://analysis.test",
        },
    )

    result = await workflows.launch_and_wait(app_id=app_id, max_wait=10)

    assert result["analysis_id"] == "analysis-456"
    assert result["url"] == "https://analysis.test"
    assert result["wait_time"] >= 0


@pytest.mark.asyncio
async def test_launch_and_wait_with_polling(
    workflows: FormationWorkflows, httpx_mock: HTTPXMock
):
    """Test launch and wait with status polling."""
    app_id = "app-123"
    analysis_id = "analysis-456"

    # Mock app config with no required parameters
    httpx_mock.add_response(
        url=f"https://formation.test/apps/de/{app_id}/config",
        json={"groups": []},
    )

    # Mock launch response without URL
    httpx_mock.add_response(
        url=f"https://formation.test/app/launch/de/{app_id}",
        json={
            "analysis_id": analysis_id,
            "name": "Test Analysis",
            "status": "Submitted",
        },
        method="POST",
    )

    # Mock first status check - not ready
    httpx_mock.add_response(
        url=f"https://formation.test/apps/analyses/{analysis_id}/status",
        json={
            "analysis_id": analysis_id,
            "status": "Running",
            "url_ready": False,
        },
    )

    # Mock second status check - ready
    httpx_mock.add_response(
        url=f"https://formation.test/apps/analyses/{analysis_id}/status",
        json={
            "analysis_id": analysis_id,
            "status": "Running",
            "url_ready": True,
            "url": "https://analysis.test",
        },
    )

    result = await workflows.launch_and_wait(
        app_id=app_id, max_wait=30, poll_interval=1
    )

    assert result["analysis_id"] == analysis_id
    assert result["url"] == "https://analysis.test"
    assert result["wait_time"] > 0


@pytest.mark.asyncio
async def test_launch_and_wait_with_required_params(
    workflows: FormationWorkflows, httpx_mock: HTTPXMock
):
    """Test launch and wait when app requires parameters."""
    app_id = "app-123"

    # Mock app config with required parameters
    httpx_mock.add_response(
        url=f"https://formation.test/apps/de/{app_id}/config",
        json={
            "groups": [
                {
                    "parameters": [
                        {
                            "id": "input_file",
                            "name": "Input File",
                            "description": "File to process",
                            "type": "FileInput",
                            "required": True,
                            "isVisible": True,
                        }
                    ]
                }
            ]
        },
    )

    result = await workflows.launch_and_wait(app_id=app_id, max_wait=10)

    assert "missing_params" in result
    assert len(result["missing_params"]) == 1
    assert result["missing_params"][0]["id"] == "input_file"
    assert result["missing_params"][0]["name"] == "Input File"


@pytest.mark.asyncio
async def test_get_running_analyses(
    workflows: FormationWorkflows, httpx_mock: HTTPXMock
):
    """Test getting running analyses."""
    httpx_mock.add_response(
        url="https://formation.test/apps/analyses/?status=Running",
        json={
            "analyses": [
                {
                    "analysis_id": "analysis-1",
                    "app_id": "app-1",
                    "system_id": "de",
                    "status": "Running",
                }
            ]
        },
    )

    result = await workflows.get_running_analyses()
    assert len(result) == 1
    assert result[0]["status"] == "Running"


def test_open_in_browser(workflows: FormationWorkflows):
    """Test opening URL in browser."""
    result = workflows.open_in_browser("https://test.com")
    # Just verify it doesn't crash - actual browser opening is platform-dependent
    assert "success" in result
    assert result["url"] == "https://test.com"
