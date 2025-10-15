"""MCP server implementation for Formation API."""

import sys

import mcp.types as types
from mcp.server import Server
from mcp.server.stdio import stdio_server

from .config import config
from .workflows import FormationWorkflows


async def _handle_list_apps(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle list_apps tool call."""
    result = await workflows.client.list_apps(
        name=arguments.get("name"),
        limit=arguments.get("limit", 10),
        offset=arguments.get("offset", 0),
    )
    # Format apps list for display
    apps = result.get("apps", [])
    if not apps:
        return [types.TextContent(type="text", text="No apps found")]

    output = f"Found {result.get('total', 0)} apps:\n\n"
    for app in apps:
        output += f"- **{app.get('name', 'Unknown')}**\n"
        output += f"  ID: `{app.get('id', 'N/A')}`\n"
        output += f"  System: {app.get('system_id', 'N/A')}\n"
        if app.get("integrator_username"):
            output += f"  Integrator: {app['integrator_username']}\n"
        if app.get("description"):
            output += f"  Description: {app['description']}\n"
        output += "\n"

    return [types.TextContent(type="text", text=output)]


async def _handle_launch_app_and_wait(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle launch_app_and_wait tool call."""
    result = await workflows.launch_and_wait(
        app_id=arguments["app_id"],
        system_id=arguments.get("system_id", "de"),
        name=arguments.get("name"),
        max_wait=arguments.get("max_wait", 300),
        config=arguments.get("config"),
        overall_job_type=arguments.get("overall_job_type"),
    )

    # Check if we need to ask for missing parameters
    if "missing_params" in result:
        output = "⚠️  This app requires additional parameters:\n\n"
        for param in result["missing_params"]:
            output += f"- **{param['name']}** (`{param['id']}`)\n"
            output += f"  Type: {param['type']}\n"
            if param.get("description"):
                output += f"  Description: {param['description']}\n"
            output += "\n"

        output += (
            "\nPlease call this tool again with a `config` parameter "
            "containing the required values.\n"
        )
        output += "Example config format:\n```json\n{\n"
        for param in result["missing_params"]:
            output += f'  "{param["id"]}": "value",\n'
        output += "}\n```"

        return [types.TextContent(type="text", text=output)]

    # Format output based on whether it's an interactive app or batch job
    url = result.get("url")

    output = "✅ Analysis launched successfully!\n\n"
    output += f"**Analysis ID:** `{result['analysis_id']}`\n"
    output += f"**Status:** {result['status']}\n"

    # Interactive apps have URLs
    if url:
        workflows.open_in_browser(url)
        output += f"**URL:** {result['url']}\n"
        output += f"**Wait time:** {result['wait_time']}s\n"
        output += "\n🌐 Opened in browser\n"
    # Batch jobs just show the job type
    elif "job_type" in result:
        output += f"**Job Type:** {result['job_type']}\n"
        output += "\n📋 Batch job submitted - check status later\n"

    return [types.TextContent(type="text", text=output)]


async def _handle_get_analysis_status(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle get_analysis_status tool call."""
    result = await workflows.client.get_analysis_status(arguments["analysis_id"])
    output = "**Analysis Status**\n\n"
    output += f"ID: `{result['analysis_id']}`\n"
    output += f"Status: {result['status']}\n"
    output += f"URL Ready: {result.get('url_ready', False)}\n"
    if result.get("url"):
        output += f"URL: {result['url']}\n"

    return [types.TextContent(type="text", text=output)]


async def _handle_list_running_analyses(
    workflows: FormationWorkflows, _arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle list_running_analyses tool call."""
    del _arguments  # Unused but required by handler signature
    analyses = await workflows.get_running_analyses()
    if not analyses:
        return [types.TextContent(type="text", text="No running analyses found")]

    output = f"Found {len(analyses)} running analyses:\n\n"
    for analysis in analyses:
        output += f"- **{analysis.get('analysis_id')}**\n"
        output += f"  App: {analysis.get('app_id')}\n"
        output += f"  System: {analysis.get('system_id')}\n"
        output += f"  Status: {analysis.get('status')}\n\n"

    return [types.TextContent(type="text", text=output)]


async def _handle_get_app_parameters(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle get_app_parameters tool call."""
    result = await workflows.client.get_app_parameters(
        arguments["app_id"],
        arguments.get("system_id", "de"),
    )

    output = f"**App Parameters: {arguments['app_id']}**\n\n"

    # Show job type if available
    if "overall_job_type" in result:
        output += f"**Job Type:** {result['overall_job_type']}\n\n"

    # Extract and display required parameters
    required_params = []
    optional_params = []

    for group in result.get("groups", []):
        for param in group.get("parameters", []):
            is_visible = param.get("isVisible", True)
            is_required = param.get("required", False)

            if is_visible and is_required:
                required_params.append(param)
            elif is_visible:
                optional_params.append(param)

    if required_params:
        output += "**Required Parameters:**\n"
        for param in required_params:
            output += f"- `{param['id']}` ({param.get('type', 'string')})\n"
            output += f"  {param.get('description', 'No description')}\n"
            if "defaultValue" in param:
                output += f"  Default: {param['defaultValue']}\n"
            output += "\n"
    else:
        output += "**No required parameters**\n\n"

    if optional_params:
        output += f"**Optional Parameters:** {len(optional_params)} available\n"

    return [types.TextContent(type="text", text=output)]


def _handle_open_in_browser(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle open_in_browser tool call."""
    result = workflows.open_in_browser(arguments["url"])
    if result["success"]:
        return [
            types.TextContent(
                type="text",
                text=f"✅ Opened {result['url']} in browser",
            )
        ]
    else:
        return [
            types.TextContent(
                type="text",
                text=f"❌ Failed to open browser: {result.get('error')}",
            )
        ]


async def _handle_stop_analysis(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle stop_analysis tool call."""
    result = await workflows.stop_analysis(
        arguments["analysis_id"],
        save_outputs=arguments.get("save_outputs", True),
    )
    del result
    save_msg = "with" if arguments.get("save_outputs", True) else "without"
    return [
        types.TextContent(
            type="text",
            text=f"✅ Analysis stopped {save_msg} saving outputs",
        )
    ]


async def _handle_create_directory(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle create_directory tool call."""
    result = await workflows.client.put_data(
        path=arguments["path"],
        resource_type="directory",
        metadata=arguments.get("metadata"),
    )

    output = f"✅ Directory created: `{result['path']}`"
    return [types.TextContent(type="text", text=output)]


async def _handle_upload_file(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle upload_file tool call."""
    content = arguments["content"]
    # Convert string content to bytes
    content_bytes = content.encode("utf-8") if isinstance(content, str) else content

    result = await workflows.client.put_data(
        path=arguments["path"],
        content=content_bytes,
        metadata=arguments.get("metadata"),
    )

    created_msg = "created" if result.get("created") else "updated"
    output = f"✅ File {created_msg}: `{result['path']}`"
    return [types.TextContent(type="text", text=output)]


async def _handle_set_metadata(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle set_metadata tool call."""
    result = await workflows.client.put_data(
        path=arguments["path"],
        metadata=arguments["metadata"],
        replace_metadata=arguments.get("replace", False),
    )

    replace_msg = "replaced" if arguments.get("replace") else "updated"
    output = f"✅ Metadata {replace_msg} for: `{result['path']}`"
    return [types.TextContent(type="text", text=output)]


async def _handle_delete_data(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle delete_data tool call."""
    result = await workflows.client.delete_data(
        path=arguments["path"],
        recurse=arguments.get("recurse", False),
        dry_run=arguments.get("dry_run", False),
    )

    # Format output based on dry_run status
    if result.get("dry_run"):
        # Dry-run output
        icon = "🔍"
        action = "Would delete" if result.get("would_delete") else "Cannot delete"
        output = f"{icon} Dry-run: {action} `{result['path']}`"
        if result.get("type") == "collection" and result.get("item_count"):
            output += f" ({result['item_count']} items)"
    else:
        # Actual deletion output
        if result.get("deleted"):
            icon = "✅"
            action = "Deleted"
            if arguments.get("recurse") and result.get("type") == "collection":
                icon = "⚠️"
                action = "Deleted (recursive)"
        else:
            icon = "❌"
            action = "Failed to delete"

        output = f"{icon} {action}: `{result['path']}`"
        if result.get("item_count"):
            output += f" ({result['item_count']} items)"

    return [types.TextContent(type="text", text=output)]


async def _handle_browse_data(
    workflows: FormationWorkflows, arguments: dict
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """Handle browse_data tool call."""
    result = await workflows.client.browse_data(
        path=arguments["path"],
        offset=arguments.get("offset", 0),
        limit=arguments.get("limit"),
        include_metadata=arguments.get("include_metadata", False),
    )

    # Check if it's a directory listing or file content
    if "type" in result and result["type"] == "collection":
        # Directory listing
        output = f"**Directory:** `{result['path']}`\n\n"
        contents = result.get("contents", [])

        if not contents:
            output += "*(empty directory)*"
        else:
            # Separate directories and files
            dirs = [item for item in contents if item["type"] == "collection"]
            files = [item for item in contents if item["type"] == "data_object"]

            if dirs:
                output += "**Directories:**\n"
                for d in dirs:
                    output += f"- 📁 {d['name']}\n"
                output += "\n"

            if files:
                output += "**Files:**\n"
                for f in files:
                    output += f"- 📄 {f['name']}\n"

        return [types.TextContent(type="text", text=output)]
    else:
        # File content
        content = result.get("content", b"")
        if isinstance(content, bytes):
            try:
                # Try to decode as text
                text_content = content.decode("utf-8")
                output = f"**File Content:**\n\n```\n{text_content}\n```"
            except UnicodeDecodeError:
                # Binary file
                output = (
                    f"**Binary File:** {len(content)} bytes\n\n"
                    f"*(Content cannot be displayed as text)*"
                )
        else:
            output = f"**File Content:**\n\n```\n{content}\n```"

        # Add metadata if present
        headers = result.get("headers", {})
        if headers:
            output += "\n\n**Metadata:**\n"
            for key, value in headers.items():
                # Remove x-datastore- prefix for display
                display_key = key.replace("x-datastore-", "").replace("-", " ").title()
                output += f"- {display_key}: {value}\n"

        return [types.TextContent(type="text", text=output)]


async def serve() -> None:
    """Run the Formation MCP server."""
    server = Server("formation")
    workflows = FormationWorkflows(
        base_url=config.base_url,
        token=config.token,
        username=config.username,
        password=config.password,
    )

    @server.list_tools()
    async def list_tools() -> list[types.Tool]:
        """List available Formation tools."""
        return [
            types.Tool(
                name="list_apps",
                description=(
                    "List available interactive VICE applications "
                    "with optional filtering by name"
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "name": {
                            "type": "string",
                            "description": "Filter apps by name (case-insensitive partial match)",
                        },
                        "limit": {
                            "type": "integer",
                            "description": "Maximum number of apps to return",
                            "default": 10,
                        },
                        "offset": {
                            "type": "integer",
                            "description": "Number of apps to skip for pagination",
                            "default": 0,
                        },
                    },
                },
            ),
            types.Tool(
                name="launch_app_and_wait",
                description=(
                    "Launch an interactive application and wait for it to become ready. "
                    "Returns the URL when ready. If the app requires parameters, "
                    "will return a list of required parameters to collect from the user."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "app_id": {
                            "type": "string",
                            "description": "UUID of the app to launch",
                        },
                        "system_id": {
                            "type": "string",
                            "description": "System identifier (default: 'de')",
                            "default": "de",
                        },
                        "name": {
                            "type": "string",
                            "description": "Custom name for the analysis (optional)",
                        },
                        "max_wait": {
                            "type": "integer",
                            "description": "Maximum seconds to wait for app to be ready",
                            "default": 300,
                        },
                        "config": {
                            "type": "object",
                            "description": "Configuration parameters required by the app",
                        },
                        "overall_job_type": {
                            "type": "string",
                            "description": (
                                "Job type from list_apps (optional). If provided from a "
                                "previous list_apps call, avoids an extra API request. "
                                "Values: 'Interactive', 'DE', 'OSG', 'Tapis'"
                            ),
                        },
                    },
                    "required": ["app_id"],
                },
            ),
            types.Tool(
                name="get_analysis_status",
                description="Check the current status of a running analysis",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "analysis_id": {
                            "type": "string",
                            "description": "UUID of the analysis",
                        },
                    },
                    "required": ["analysis_id"],
                },
            ),
            types.Tool(
                name="list_running_analyses",
                description="List all currently running analyses for the authenticated user",
                inputSchema={
                    "type": "object",
                    "properties": {},
                },
            ),
            types.Tool(
                name="get_app_parameters",
                description=(
                    "Get the parameters for a specific app, including required parameters, "
                    "parameter types, and default values. Use this to check what parameters "
                    "an app needs before launching it."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "app_id": {
                            "type": "string",
                            "description": "UUID of the app",
                        },
                        "system_id": {
                            "type": "string",
                            "description": "System identifier (default: 'de')",
                            "default": "de",
                        },
                    },
                    "required": ["app_id"],
                },
            ),
            types.Tool(
                name="open_in_browser",
                description="Open an analysis URL in the default web browser",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "url": {
                            "type": "string",
                            "description": "URL to open in browser",
                        },
                    },
                    "required": ["url"],
                },
            ),
            types.Tool(
                name="stop_analysis",
                description="Stop a running analysis, optionally saving outputs",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "analysis_id": {
                            "type": "string",
                            "description": "UUID of the analysis to stop",
                        },
                        "save_outputs": {
                            "type": "boolean",
                            "description": "Whether to save outputs before stopping",
                            "default": True,
                        },
                    },
                    "required": ["analysis_id"],
                },
            ),
            types.Tool(
                name="browse_data",
                description=(
                    "Browse iRODS data store directory or read file contents. "
                    "For directories, returns a list of contents. "
                    "For files, returns the file content."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": (
                                "Full iRODS path to browse (e.g., "
                                "'/iplant/home/username/directory' or "
                                "'/iplant/home/username/file.txt')"
                            ),
                        },
                        "offset": {
                            "type": "integer",
                            "description": "Byte offset for file reading (default: 0)",
                            "default": 0,
                        },
                        "limit": {
                            "type": "integer",
                            "description": "Max bytes to read for files (optional)",
                        },
                        "include_metadata": {
                            "type": "boolean",
                            "description": "Include iRODS AVU metadata (default: false)",
                            "default": False,
                        },
                    },
                    "required": ["path"],
                },
            ),
            types.Tool(
                name="create_directory",
                description="Create a new directory in the iRODS data store",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": (
                                "Full iRODS path for the new directory "
                                "(e.g., '/iplant/home/username/newdir')"
                            ),
                        },
                        "metadata": {
                            "type": "object",
                            "description": (
                                "Optional metadata as key-value pairs "
                                "(e.g., {'author': 'username', 'project': 'myproject'})"
                            ),
                        },
                    },
                    "required": ["path"],
                },
            ),
            types.Tool(
                name="upload_file",
                description="Upload a file to the iRODS data store",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": (
                                "Full iRODS path for the file "
                                "(e.g., '/iplant/home/username/file.txt')"
                            ),
                        },
                        "content": {
                            "type": "string",
                            "description": "File content as a string",
                        },
                        "metadata": {
                            "type": "object",
                            "description": (
                                "Optional metadata as key-value pairs "
                                "(e.g., {'author': 'username', 'filetype': 'text'})"
                            ),
                        },
                    },
                    "required": ["path", "content"],
                },
            ),
            types.Tool(
                name="set_metadata",
                description="Set or update metadata on an existing file or directory",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": (
                                "Full iRODS path to file or directory "
                                "(e.g., '/iplant/home/username/file.txt')"
                            ),
                        },
                        "metadata": {
                            "type": "object",
                            "description": (
                                "Metadata as key-value pairs "
                                "(e.g., {'author': 'username', 'version': '1.0'})"
                            ),
                        },
                        "replace": {
                            "type": "boolean",
                            "description": (
                                "If true, replace all existing metadata. "
                                "If false, add to existing metadata (default: false)"
                            ),
                            "default": False,
                        },
                    },
                    "required": ["path", "metadata"],
                },
            ),
            types.Tool(
                name="delete_data",
                description=(
                    "Delete a file or directory from the iRODS data store. "
                    "Supports dry-run mode to preview deletion without executing. "
                    "**Safety:** Deletions are permanent. Consider using dry_run=true first."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": (
                                "Full iRODS path to delete "
                                "(e.g., '/iplant/home/username/file.txt' or "
                                "'/iplant/home/username/directory')"
                            ),
                        },
                        "recurse": {
                            "type": "boolean",
                            "description": (
                                "Delete non-empty directories. "
                                "Default is false for safety."
                            ),
                            "default": False,
                        },
                        "dry_run": {
                            "type": "boolean",
                            "description": (
                                "Preview what would be deleted without actually deleting. "
                                "Highly recommended before actual deletion."
                            ),
                            "default": False,
                        },
                    },
                    "required": ["path"],
                },
            ),
        ]

    @server.call_tool()
    async def call_tool(
        name: str, arguments: dict
    ) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
        """Handle tool calls from the MCP client."""
        try:
            if name == "list_apps":
                return await _handle_list_apps(workflows, arguments)
            elif name == "launch_app_and_wait":
                return await _handle_launch_app_and_wait(workflows, arguments)
            elif name == "get_analysis_status":
                return await _handle_get_analysis_status(workflows, arguments)
            elif name == "list_running_analyses":
                return await _handle_list_running_analyses(workflows, arguments)
            elif name == "get_app_parameters":
                return await _handle_get_app_parameters(workflows, arguments)
            elif name == "open_in_browser":
                return _handle_open_in_browser(workflows, arguments)
            elif name == "stop_analysis":
                return await _handle_stop_analysis(workflows, arguments)
            elif name == "browse_data":
                return await _handle_browse_data(workflows, arguments)
            elif name == "create_directory":
                return await _handle_create_directory(workflows, arguments)
            elif name == "upload_file":
                return await _handle_upload_file(workflows, arguments)
            elif name == "set_metadata":
                return await _handle_set_metadata(workflows, arguments)
            elif name == "delete_data":
                return await _handle_delete_data(workflows, arguments)
            else:
                return [types.TextContent(type="text", text=f"Unknown tool: {name}")]

        except Exception as e:
            import traceback

            error_msg = f"❌ Error executing {name}: {str(e)}"
            print(error_msg, file=sys.stderr)
            print(traceback.format_exc(), file=sys.stderr)
            return [types.TextContent(type="text", text=error_msg)]

    # Run the server
    async with stdio_server() as (read_stream, write_stream):
        await server.run(read_stream, write_stream, server.create_initialization_options())
