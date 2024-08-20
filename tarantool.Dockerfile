FROM tarantool/tarantool:latest

WORKDIR /opt/tarantool

COPY ./tarantInit.lua ./init.lua

ENV TT_USERNAME storage

EXPOSE 3301

CMD ["tarantool", "init.lua"]
