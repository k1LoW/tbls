require 'erb'

n = 1000

contents = <<EOS
<% (1..n).each do |i| %>
CREATE TABLE t<%= i %> (
  id serial PRIMARY KEY,
<% if i > 1 then %>
  t<%= i - 1 %>_id int NOT NULL,
<% end %>
  created timestamp NOT NULL,
  updated timestamp<% if i > 1 then %>,
  CONSTRAINT t<%= i - 1 %>_id_fk FOREIGN KEY(t<%= i - 1 %>_id) REFERENCES t<%= i - 1 %>(id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE CASCADE<% end %>
);
<% end %>
EOS

erb = ERB.new(contents)
puts erb.result(binding)
