create table comments(
    telegram_id bigint primary key not null,
    text TEXT not null,
    user_id bigint not null,

    created_at timestamp not null default NOW(),
    updated_at timestamp not null default NOW()
);

alter table comments add column chat_id bigint not null default 0;