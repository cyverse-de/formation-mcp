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
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
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
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
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
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
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
        url=f"https://formation.test/data/{test_path.lstrip('/')}?limit=5",
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
        url=f"https://formation.test/data/{test_path.lstrip('/')}?include_metadata=true",
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
        url=f"https://formation.test/data/{test_path.lstrip('/')}?include_metadata=true",
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


@pytest.mark.asyncio
async def test_create_directory(client: FormationClient, httpx_mock: HTTPXMock):
    """Test creating a directory."""
    test_path = "/iplant/home/testuser/newdir"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}?resource_type=directory",
        json={
            "path": test_path,
            "type": "collection",
            "created": True,
        },
        method="PUT",
    )

    result = await client.put_data(test_path, resource_type="directory")

    assert result["path"] == test_path
    assert result["type"] == "collection"
    assert result["created"] is True


@pytest.mark.asyncio
async def test_upload_file(client: FormationClient, httpx_mock: HTTPXMock):
    """Test uploading a file."""
    test_path = "/iplant/home/testuser/newfile.txt"
    content = b"Test file content"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "data_object",
            "created": True,
        },
        method="PUT",
    )

    result = await client.put_data(test_path, content=content)

    assert result["path"] == test_path
    assert result["type"] == "data_object"
    assert result["created"] is True


@pytest.mark.asyncio
async def test_upload_file_with_metadata(client: FormationClient, httpx_mock: HTTPXMock):
    """Test uploading a file with metadata."""
    test_path = "/iplant/home/testuser/newfile.txt"
    content = b"Test file content"
    metadata = {"author": "testuser", "project": "test"}

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "data_object",
            "created": True,
        },
        method="PUT",
    )

    result = await client.put_data(test_path, content=content, metadata=metadata)

    assert result["path"] == test_path
    assert result["created"] is True


@pytest.mark.asyncio
async def test_set_metadata(client: FormationClient, httpx_mock: HTTPXMock):
    """Test setting metadata on an existing path."""
    test_path = "/iplant/home/testuser/file.txt"
    metadata = {"author": "testuser", "version": "1.0"}

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "data_object",
            "created": False,
        },
        method="PUT",
    )

    result = await client.put_data(test_path, metadata=metadata)

    assert result["path"] == test_path
    assert result["created"] is False


@pytest.mark.asyncio
async def test_replace_metadata(client: FormationClient, httpx_mock: HTTPXMock):
    """Test replacing metadata on an existing path."""
    test_path = "/iplant/home/testuser/file.txt"
    metadata = {"author": "newuser"}

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}?replace_metadata=true",
        json={
            "path": test_path,
            "type": "data_object",
            "created": False,
        },
        method="PUT",
    )

    result = await client.put_data(test_path, metadata=metadata, replace_metadata=True)

    assert result["path"] == test_path
    assert result["created"] is False


@pytest.mark.asyncio
async def test_delete_file(client: FormationClient, httpx_mock: HTTPXMock):
    """Test deleting a file."""
    test_path = "/iplant/home/testuser/file.txt"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "data_object",
            "would_delete": True,
            "deleted": True,
            "dry_run": False,
        },
        method="DELETE",
    )

    result = await client.delete_data(test_path)

    assert result["path"] == test_path
    assert result["deleted"] is True
    assert result["dry_run"] is False


@pytest.mark.asyncio
async def test_delete_file_dry_run(client: FormationClient, httpx_mock: HTTPXMock):
    """Test dry-run file deletion."""
    test_path = "/iplant/home/testuser/file.txt"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}?dry_run=true",
        json={
            "path": test_path,
            "type": "data_object",
            "would_delete": True,
            "deleted": False,
            "dry_run": True,
        },
        method="DELETE",
    )

    result = await client.delete_data(test_path, dry_run=True)

    assert result["path"] == test_path
    assert result["would_delete"] is True
    assert result["deleted"] is False
    assert result["dry_run"] is True


@pytest.mark.asyncio
async def test_delete_empty_directory(client: FormationClient, httpx_mock: HTTPXMock):
    """Test deleting an empty directory."""
    test_path = "/iplant/home/testuser/emptydir"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}",
        json={
            "path": test_path,
            "type": "collection",
            "would_delete": True,
            "deleted": True,
            "dry_run": False,
        },
        method="DELETE",
    )

    result = await client.delete_data(test_path)

    assert result["path"] == test_path
    assert result["deleted"] is True


@pytest.mark.asyncio
async def test_delete_directory_with_recurse(
    client: FormationClient, httpx_mock: HTTPXMock
):
    """Test deleting a non-empty directory with recurse."""
    test_path = "/iplant/home/testuser/folder"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}?recurse=true",
        json={
            "path": test_path,
            "type": "collection",
            "would_delete": True,
            "deleted": True,
            "dry_run": False,
            "item_count": 15,
        },
        method="DELETE",
    )

    result = await client.delete_data(test_path, recurse=True)

    assert result["path"] == test_path
    assert result["deleted"] is True
    assert result["item_count"] == 15


@pytest.mark.asyncio
async def test_delete_directory_dry_run_with_recurse(
    client: FormationClient, httpx_mock: HTTPXMock
):
    """Test dry-run directory deletion with recurse."""
    test_path = "/iplant/home/testuser/folder"

    httpx_mock.add_response(
        url=f"https://formation.test/data/{test_path.lstrip('/')}?recurse=true&dry_run=true",
        json={
            "path": test_path,
            "type": "collection",
            "would_delete": True,
            "deleted": False,
            "dry_run": True,
            "item_count": 15,
        },
        method="DELETE",
    )

    result = await client.delete_data(test_path, recurse=True, dry_run=True)

    assert result["path"] == test_path
    assert result["would_delete"] is True
    assert result["deleted"] is False
    assert result["dry_run"] is True
    assert result["item_count"] == 15
