

`CAT_publish`: even not existing or false -> CAT won't publish in the proceedings any contribution having CAT_publish set to false


`duplicate_of`: a string with the programme code of the parent contribution, i.e. the contribution this one was cloned from. e.g.: SUPM001 has duplicate_of = WEPL135
In other words, any contribution having a not-null/empty duplicate_of is not to be indexed in the refdb (the proceedings report the parent contribution as reference)