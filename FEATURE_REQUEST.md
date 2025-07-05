# Feature Request: Add Logical Name Column Support for Table Documentation

## Summary
Add support for displaying logical names (Japanese/local language names) as a separate column in table documentation, extracted from database column comments.

## Requirements

### 1. Logical Name Column Display
- Add a new "Logical Name" column in table documentation output
- Display alongside the existing Name column

### 2. Configuration Control
- Enable/disable logical name column output via tbls configuration file
- Should be configurable per project

### 3. Data Source
- Extract logical name data from database table column comments
- Support all databases that tbls currently supports (PostgreSQL, MySQL, SQLite, etc.)

### 4. Fallback Behavior
- When logical name data doesn't exist, display the physical column name (Name) as fallback
- Ensure no empty cells in the logical name column

### 5. Column Order (when logical name enabled)
When logical name column is enabled, the table column order should be:
1. Name (physical column name)
2. Logical Name (new column)
3. Type 
4. Default
5. Nullable
6. Children
7. Parents  
8. Comment

### 6. Comment Field Processing
When logical name column is enabled, process the Comment field content:
- Split comment content using a delimiter
- First part → Logical Name column
- Second part → Comment column

### 7. Comment Display
When logical name column is enabled:
- Comment column should only display content after the delimiter
- Hide the logical name portion from the comment field

### 8. Delimiter Configuration
- Default delimiter: `|` (pipe character)
- Allow customization via configuration file
- Example comment format: `ユーザーID|Unique identifier for users`

## Configuration Example

```yaml
# .tbls.yml
format:
  logicalName:
    enabled: true
    delimiter: "|"
    fallbackToName: true
```

## Use Cases

### Current Output
| Name     | Type         | Default | Nullable | Comment           |
|----------|--------------|---------|----------|-------------------|
| user_id  | integer      |         | false    | ユーザーID            |
| username | varchar(50)  |         | false    | ユーザー名            |

### Expected Output (with logical name enabled)
| Name     | Logical Name | Type         | Default | Nullable | Comment |
|----------|--------------|--------------|---------|----------|---------|
| user_id  | ユーザーID       | integer      |         | false    |         |
| username | ユーザー名       | varchar(50)  |         | false    |         |

### With Delimiter Comments
Database comment: `ユーザーID|Unique identifier for users`

| Name     | Logical Name | Type         | Default | Nullable | Comment                        |
|----------|--------------|--------------|---------|----------|--------------------------------|
| user_id  | ユーザーID       | integer      |         | false    | Unique identifier for users    |

## Implementation Considerations

1. **Backward Compatibility**: This feature should be opt-in and not affect existing functionality
2. **Template Updates**: Modify existing templates (MD, HTML, etc.) to support the new column
3. **Multi-language Support**: Should work with any language, not just Japanese
4. **Performance**: Should not significantly impact documentation generation performance
5. **Testing**: Comprehensive tests for all supported database types

## Related Files to Modify
- `config/config.go` - Add configuration options
- `schema/schema.go` - Extend Column struct if needed
- `output/md/templates/table.md.tmpl` - Update template
- `output/*/templates/` - Update all format templates
- Tests for all supported databases

## Benefits
- Better documentation for international teams
- Separation of technical and business terminology
- Improved readability for non-technical stakeholders
- Flexible comment field usage

## Priority
Medium - This would be a valuable addition for teams working with multi-language documentation requirements.