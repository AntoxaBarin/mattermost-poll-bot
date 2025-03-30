box.cfg{
    listen = 3301,
}

box.schema.user.grant('guest', 'read,write,execute', 'universe')

box.schema.space.create('polls', {
    format = {
        {name = 'id', type = 'unsigned'},
        {name = 'title', type = 'string'},
        {name = 'options', type = 'map'},
        {name = 'timestamp', type = 'string'},
        {name = 'creator', type = 'string'},
    }
})
box.space.polls:create_index('primary', {parts = {'id'}})
