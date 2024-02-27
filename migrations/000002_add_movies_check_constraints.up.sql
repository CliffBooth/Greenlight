alter table movies add constraint movies_runtime_check check (runtime >= 0);

alter table movies add constraint movies_year_check check (year > 1887);

alter table movies add constraint genres_length_check check (array_length(genres, 1) >= 0);