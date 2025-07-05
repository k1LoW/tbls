-- usersテーブルのカラムに論理名付きコメントを追加
COMMENT ON TABLE users IS 'ユーザー情報を管理するマスターテーブル';
COMMENT ON COLUMN users.id IS 'ユーザーID|システム内で一意のユーザー識別子';
COMMENT ON COLUMN users.username IS 'ユーザー名|ログイン時に使用する名前（4文字以上必須）';
COMMENT ON COLUMN users.password IS 'パスワード|ハッシュ化されたパスワード';
COMMENT ON COLUMN users.email IS 'メールアドレス|連絡用メールアドレス（一意制約あり）';
COMMENT ON COLUMN users.created IS '登録日時|アカウント作成日時';
COMMENT ON COLUMN users.updated IS '更新日時|最終更新日時';

-- postsテーブルのカラムに論理名付きコメントを追加
COMMENT ON TABLE posts IS 'ブログ投稿を管理するテーブル';
COMMENT ON COLUMN posts.id IS '投稿ID|投稿の一意識別子';
COMMENT ON COLUMN posts.user_id IS '投稿者ID|投稿したユーザーのID（外部キー）';
COMMENT ON COLUMN posts.title IS 'タイトル|投稿のタイトル（最大255文字）';
COMMENT ON COLUMN posts.body IS '本文|投稿の本文内容';
COMMENT ON COLUMN posts.created IS '投稿日時|投稿作成日時';
COMMENT ON COLUMN posts.updated IS '更新日時|投稿の最終更新日時';

-- commentsテーブルのカラムに論理名付きコメントを追加
COMMENT ON TABLE comments IS '投稿へのコメントを管理するテーブル';
COMMENT ON COLUMN comments.id IS 'コメントID|コメントの一意識別子';
COMMENT ON COLUMN comments.post_id IS '投稿ID|コメント対象の投稿ID（外部キー）';
COMMENT ON COLUMN comments.user_id IS 'コメント者ID|コメントしたユーザーのID（外部キー）';
COMMENT ON COLUMN comments.comment IS 'コメント内容|コメントの本文';
COMMENT ON COLUMN comments.created IS 'コメント日時|コメント投稿日時';
COMMENT ON COLUMN comments.updated IS '更新日時|コメントの最終更新日時';