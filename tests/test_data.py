"""Tests for data browsing functionality."""

import pytest
from pytest_httpx import HTTPXMock

from formation_mcp.client import FormationClient


@pytest.fixture
def client():
    """Create test client instance."""
    return FormationClient(
        base_url="https://formation.test",
        token="test-token",
    )


@pytest.mark.asyncio
async def test_browse_directory(client: FormationClient, httpx_mock: HTTPXMock):
    """Test browsing a directory."""
    test_path = "/iplant/home/testuser"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "collection",
            "contents": [
                {"name": "subdir", "type": "collection"},
                {"name": "file.txt", "type": "data_object"},
            ],
        },
        headers={"content-type": "application/json"},
    )

    result = await client.browse_data(test_path)

    assert result["type"] == "collection"
    assert result["path"] == test_path
    assert len(result["contents"]) == 2
    assert result["contents"][0]["name"] == "subdir"
    assert result["contents"][1]["name"] == "file.txt"


@pytest.mark.asyncio
async def test_browse_empty_directory(client: FormationClient, httpx_mock: HTTPXMock):
    """Test browsing an empty directory."""
    test_path = "/iplant/home/testuser/empty"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "collection",
            "contents": [],
        },
        headers={"content-type": "application/json"},
    )

    result = await client.browse_data(test_path)

    assert result["type"] == "collection"
    assert result["contents"] == []


@pytest.mark.asyncio
async def test_read_text_file(client: FormationClient, httpx_mock: HTTPXMock):
    """Test reading a text file."""
    test_path = "/iplant/home/testuser/file.txt"
    file_content = "Hello, world!"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}",
        content=file_content.encode("utf-8"),
        headers={"content-type": "text/plain"},
    )

    result = await client.browse_data(test_path)

    assert "content" in result
    assert result["content"] == file_content.encode("utf-8")


@pytest.mark.asyncio
async def test_read_file_with_offset_and_limit(
    client: FormationClient, httpx_mock: HTTPXMock
):
    """Test reading a file with offset and limit."""
    test_path = "/iplant/home/testuser/file.txt"
    file_content = "Hello, world!"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}?limit=5",
        content=file_content.encode("utf-8"),
        headers={"content-type": "text/plain"},
    )

    result = await client.browse_data(test_path, offset=0, limit=5)

    assert "content" in result


@pytest.mark.asyncio
async def test_browse_with_metadata(client: FormationClient, httpx_mock: HTTPXMock):
    """Test browsing with metadata included."""
    test_path = "/iplant/home/testuser"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}?include_metadata=true",
        json={
            "path": test_path,
            "type": "collection",
            "contents": [],
        },
        headers={
            "content-type": "application/json",
            "X-Datastore-Author": "testuser",
            "X-Datastore-Created": "2024-01-01",
        },
    )

    result = await client.browse_data(test_path, include_metadata=True)

    assert result["type"] == "collection"


@pytest.mark.asyncio
async def test_read_file_with_metadata(client: FormationClient, httpx_mock: HTTPXMock):
    """Test reading a file with metadata."""
    test_path = "/iplant/home/testuser/file.txt"
    file_content = "Hello, world!"

    httpx_mock.add_response(
        url=f"https://formation.test/data/browse/{test_path.lstrip('/')}?include_metadata=true",
        content=file_content.encode("utf-8"),
        headers={
            "content-type": "text/plain",
            "X-Datastore-Size": "13",
            "X-Datastore-Modified": "2024-01-01",
        },
    )

    result = await client.browse_data(test_path, include_metadata=True)

    assert "content" in result
    assert "headers" in result
    assert "x-datastore-size" in result["headers"]
    assert result["headers"]["x-datastore-size"] == "13"
