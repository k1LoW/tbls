package ddl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseReferencedTables(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{
			`SELECT * FROM posts`,
			[]string{"posts"},
		},
		{
			`CREATE VIEW post_comments AS (
 SELECT c.id,
    p.title,
    u.username AS post_user,
    c.comment,
    u2.username AS comment_user,
    c.created,
    c.updated
   FROM (((posts p
     LEFT JOIN comments c ON ((p.id = c.post_id)))
     LEFT JOIN users u ON ((u.id = p.user_id)))
     LEFT JOIN users u2 ON ((u2.id = c.user_id)))
)`,
			[]string{"posts", "comments", "users"},
		},
		{
			"CREATE VIEW post_comments AS (select `c`.`id` AS `id`,`p`.`title` AS `title`,`u2`.`username` AS `post_user`,`c`.`comment` AS `comment`,`u2`.`username` AS `comment_user`,`c`.`created` AS `created`,`c`.`updated` AS `updated` from (((`testdb`.`posts` `p` left join `testdb`.`comments` `c` on((`p`.`id` = `c`.`post_id`))) left join `testdb`.`users` `u` on((`u`.`id` = `p`.`user_id`))) left join `testdb`.`users` `u2` on((`u2`.`id` = `c`.`user_id`))))",
			[]string{"testdb.posts", "testdb.comments", "testdb.users"},
		},
		{
			"CREATE VIEW k1low_posts AS (SELECT * FROM posts WHERE user_id IN (SELECT id FROM users WHERE email = 'k1lowxb@gmail.com'))",
			[]string{"posts", "users"},
		},
		{
			`CREATE VIEW k1low_posts AS (
WITH k1low AS (SELECT * FROM users WHERE email = 'k1lowxb@gmail.com')
SELECT * FROM posts WHERE user_id IN (SELECT id FROM k1low)
)
`,
			[]string{"users", "posts"},
		},
	}
	for _, tt := range tests {
		got := ParseReferencedTables(tt.in)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
