create table if not exists images (
  id serial primary key,
	user_id uuid references auth.users,
	created_at timestamp not null default now(),
  status int not null default 1,
  prompt text not null,
  batch_id uuid not null,
  image_location text,
  deleted boolean not null default 'false',
  deleted_at timestamp
)