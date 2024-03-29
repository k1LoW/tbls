# testdb

## Description

Sample database document.

## Labels

`sample` `tbls`

## Viewpoints

| Name | Description |
| ---- | ----------- |
| [Content](viewpoint-0.md) | Content as an asset for blogging services |
| [Ops](viewpoint-1.md) | Tables to be referenced during operation |
| [Around the users table](viewpoint-2.md) | Tables related to the users table |
| [Secure data](viewpoint-3.md) | Tables with secure data |

## Tables

| Name | Columns | Comment | Type | Labels |
| ---- | ------- | ------- | ---- | ------ |
| [CamelizeTable](CamelizeTable.md) | 2 |  | BASE TABLE |  |
| [comment_stars](comment_stars.md) | 6 |  | BASE TABLE | `content` |
| [comments](comments.md) | 7 | Comments<br>Multi-line<br>table<br>comment | BASE TABLE | `content` |
| [hyphen-table](hyphen-table.md) | 3 |  | BASE TABLE |  |
| [logs](logs.md) | 7 | Auditログ | BASE TABLE |  |
| [long_long_long_long_long_long_long_long_table_name](long_long_long_long_long_long_long_long_table_name.md) | 2 |  | BASE TABLE |  |
| [post_comments](post_comments.md) | 7 | post and comments View table | VIEW | `content` |
| [posts](posts.md) | 7 | Posts table | BASE TABLE | `content` |
| [user_options](user_options.md) | 4 | User options table | BASE TABLE | `user` |
| [users](users.md) | 6 | Users table | BASE TABLE | `user` |

## Stored procedures and functions

| Name | ReturnType | Arguments | Type |
| ---- | ------- | ------- | ---- |
| CustomerLevel | varchar | credit decimal | FUNCTION |
| GetAllComments |  |  | PROCEDURE |

## Relations

![er](schema.svg)

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
