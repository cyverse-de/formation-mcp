# Formation MCP API Coverage Analysis

## Summary of Missing Parameters

### 1. list_apps Tool
**Formation API `/apps` supports:**
- ✅ limit
- ✅ offset
- ✅ name
- ✅ integrator
- ❌ **description** - Filter by app description
- ❌ **integration_date** - Filter by integration date
- ❌ **edited_date** - Filter by edit date
- ❌ **job_type** - Filter by job type (Interactive, DE, OSG, Tapis)

### 2. get_app_parameters Tool
**Formation API `/apps/{system_id}/{app_id}/parameters` supports:**
- ✅ system_id
- ✅ app_id
- ✅ All parameters covered

### 3. launch_app_and_wait Tool
**Formation API `/app/launch/{system_id}/{app_id}` supports:**
- ✅ system_id
- ✅ app_id
- ✅ submission body (name, config, etc.)
- ❌ **overall_job_type** - This parameter doesn't exist in Formation API (should be removed)
- ✅ output_zone - Not exposed in MCP but handled by Formation defaults

### 4. get_analysis_status Tool
**Formation API `/apps/analyses/{analysis_id}/status` supports:**
- ✅ analysis_id
- ✅ All parameters covered

### 5. list_running_analyses Tool
**Formation API `/apps/analyses/` supports:**
- ❌ **status** - Currently hardcoded to "Running", should be a parameter
  - Possible values: "Running", "Completed", "Failed", "Submitted", "Canceled"

### 6. stop_analysis Tool
**Formation API `/apps/analyses/{analysis_id}/control` supports:**
- ✅ analysis_id
- ✅ operation (via query parameter)
- ✅ save_outputs (via query parameter)
- ✅ All parameters covered

### 7. browse_data Tool
**Formation API `/data/{path}` GET supports:**
- ✅ path
- ✅ offset
- ✅ limit
- ✅ include_metadata
- ❌ **avu_delimiter** - Delimiter for parsing AVU metadata (default: ",")

### 8. create_directory Tool
**Formation API `/data/{path}` PUT supports:**
- ✅ path
- ✅ resource_type=directory (query parameter)
- ✅ metadata (via X-Datastore-* headers)
- ❌ **avu_delimiter** - Delimiter for metadata values/units (default: ",")
- ❌ **replace_metadata** - Whether to replace existing metadata (default: false)

### 9. upload_file Tool
**Formation API `/data/{path}` PUT supports:**
- ✅ path
- ✅ content (request body)
- ✅ metadata (via X-Datastore-* headers)
- ❌ **avu_delimiter** - Delimiter for metadata values/units (default: ",")
- ❌ **replace_metadata** - Whether to replace existing metadata (default: false)

### 10. set_metadata Tool
**Formation API `/data/{path}` PUT supports:**
- ✅ path
- ✅ metadata (via X-Datastore-* headers)
- ❌ **avu_delimiter** - Delimiter for metadata values/units (default: ",")
- ❌ **replace_metadata** - Whether to replace existing metadata (default: false)

### 11. delete_data Tool
**Formation API `/data/{path}` DELETE supports:**
- ✅ path
- ✅ dry_run
- ✅ recurse
- ✅ All parameters covered

### 12. open_in_browser Tool
**Local operation - not an API call**
- ✅ url
- ✅ All parameters covered

## Priority Issues

### COMPLETED ✅
1. ✅ **Removed overall_job_type** from launch_app_and_wait - didn't exist in API
2. ✅ **Added status parameter** to list_running_analyses - was hardcoded to "Running"
3. ✅ **Fixed replace_metadata** parameter name in set_metadata - now uses correct `replace_metadata` query param
4. ✅ **Added description filter** to list_apps
5. ✅ **Added job_type filter** to list_apps

### REMAINING - MEDIUM PRIORITY
6. **Add avu_delimiter** to browse_data, set_metadata, upload_file, create_directory
   - This controls how metadata values/units are parsed (default: ",")
   - Lower priority as defaults work for most cases

### REMAINING - LOW PRIORITY
7. **Add integration_date filter** to list_apps
8. **Add edited_date filter** to list_apps
   - These are specialized filters not commonly used

## Notes
- The Formation API uses X-Datastore-* headers for metadata
- The avu_delimiter parameter controls how value/units are separated in metadata headers (default ",")
- replace_metadata=true will replace all existing metadata, false will add/update specific attributes
