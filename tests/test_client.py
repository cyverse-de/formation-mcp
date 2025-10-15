"""Tests for Formation API client."""

import pytest
from pytest_httpx import HTTPXMock

from formation_mcp.client import FormationClient


@pytest.fixture
def client():
    """Create a test client with token."""
    return FormationClient(
        base_url="https://formation.test",
        token="test-token",
    )


@pytest.fixture
def client_with_credentials():
    """Create a test client with username/password."""
    return FormationClient(
        base_url="https://formation.test",
        username="testuser",
        password="testpass",
    )


@pytest.mark.asyncio
async def test_list_apps(client: FormationClient, httpx_mock: HTTPXMock):
    """Test listing apps."""
    httpx_mock.add_response(
        url="https://formation.test/apps?limit=10&offset=0",
        json={
            "total": 1,
            "apps": [
                {
                    "id": "app-123",
                    "name": "Test App",
                    "system_id": "de",
                    "description": "A test app",
                }
            ],
        },
    )

    result = await client.list_apps(limit=10, offset=0)
    assert result["total"] == 1
    assert len(result["apps"]) == 1
    assert result["apps"][0]["name"] == "Test App"


@pytest.mark.asyncio
async def test_launch_app(client: FormationClient, httpx_mock: HTTPXMock):
    """Test launching an app."""
    app_id = "app-123"

    httpx_mock.add_response(
        url=f"https://formation.test/app/launch/de/{app_id}",
        json={
            "analysis_id": "analysis-456",
            "name": "Test Analysis",
            "status": "Submitted",
        },
        method="POST",
    )

    result = await client.launch_app(app_id=app_id, name="Test Analysis")
    assert result["analysis_id"] == "analysis-456"
    assert result["status"] == "Submitted"


@pytest.mark.asyncio
async def test_get_app_parameters(client: FormationClient, httpx_mock: HTTPXMock):
    """Test getting app parameters."""
    app_id = "app-123"

    httpx_mock.add_response(
        url=f"https://formation.test/apps/de/{app_id}/parameters",
        json={
            "overall_job_type": "DE",
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
            ],
        },
    )

    result = await client.get_app_parameters(app_id)
    assert result["overall_job_type"] == "DE"
    assert len(result["groups"]) == 1
    assert len(result["groups"][0]["parameters"]) == 1


@pytest.mark.asyncio
async def test_get_analysis_status(client: FormationClient, httpx_mock: HTTPXMock):
    """Test getting analysis status."""
    analysis_id = "analysis-456"

    httpx_mock.add_response(
        url=f"https://formation.test/apps/analyses/{analysis_id}/status",
        json={
            "analysis_id": analysis_id,
            "status": "Running",
            "url_ready": True,
            "url": "https://analysis.test",
        },
    )

    result = await client.get_analysis_status(analysis_id)
    assert result["status"] == "Running"
    assert result["url_ready"] is True
    assert result["url"] == "https://analysis.test"


@pytest.mark.asyncio
async def test_list_analyses(client: FormationClient, httpx_mock: HTTPXMock):
    """Test listing analyses."""
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

    result = await client.list_analyses(status="Running")
    assert len(result["analyses"]) == 1
    assert result["analyses"][0]["status"] == "Running"


@pytest.mark.asyncio
async def test_login_success(
    client_with_credentials: FormationClient, httpx_mock: HTTPXMock
):
    """Test successful login with username/password."""
    httpx_mock.add_response(
        url="https://formation.test/login",
        json={
            "access_token": "new-token-123",
            "token_type": "bearer",
            "expires_in": 3600,
        },
        method="POST",
    )

    result = await client_with_credentials.login("testuser", "testpass")
    assert result["access_token"] == "new-token-123"
    assert result["token_type"] == "bearer"
    assert result["expires_in"] == 3600


@pytest.mark.asyncio
async def test_auto_login_on_first_request(
    client_with_credentials: FormationClient, httpx_mock: HTTPXMock
):
    """Test that first API call triggers automatic login."""
    # Mock login endpoint
    httpx_mock.add_response(
        url="https://formation.test/login",
        json={
            "access_token": "auto-token-456",
            "token_type": "bearer",
            "expires_in": 3600,
        },
        method="POST",
    )

    # Mock apps endpoint
    httpx_mock.add_response(
        url="https://formation.test/apps?limit=5&offset=0",
        json={"total": 0, "apps": []},
    )

    # First API call should trigger login
    result = await client_with_credentials.list_apps(limit=5)
    assert result["total"] == 0

    # Verify login was called
    login_requests = [
        r for r in httpx_mock.get_requests() if r.url.path == "/login"
    ]
    assert len(login_requests) == 1
