"""Entry point for the formation-mcp server."""

import asyncio
import sys

from .server import serve


def main() -> None:
    """Main entry point for the MCP server."""
    try:
        asyncio.run(serve())
    except KeyboardInterrupt:
        print("\nShutting down Formation MCP server...", file=sys.stderr)
        sys.exit(0)
    except Exception as e:
        print(f"Fatal error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
