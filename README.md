Order Processing System
------

### How to start Order Processing Service?
> cd order_process/

> go build

> ./order_process

### How to qurey the status of Order Processing Service?

> curl -H "Authorization:user" http://localhost:8080/diagnostic/heartbeat

        {"serive_id":"f3bf183f-76b6-45eb-74d5-a970adfcfa99","status":"OK"}


### How to submit new order?

> curl -X POST --data "{}" -H "Authorization:user" http://127.0.0.1:8080/orders

        {"order_id":"8cc227c0-8dac-42cf-783e-f7bcb95bf455","start_time":"2016-03-27 10:22:31.4492618 +0000 UTC"}


### How to qurey the order state?

> curl -H "Authorization:user" http://localhost:8080/orders/8cc227c0-8dac-42cf-783e-f7bcb95bf455

        {
            "complete_time": "2016-03-27T10:22:51.4504397Z",
            "curent_step": "Completed",
            "finished": true,
            "order_id": "8cc227c0-8dac-42cf-783e-f7bcb95bf455",
            "start_time": "2016-03-27T10:22:31.4492618Z",
            "steps": [
                {
                    "step_complete_time": "2016-03-27T10:22:36.4500059Z",
                    "step_name": "Scheduling",
                    "step_start_time": "2016-03-27T10:22:31.4492618Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:41.4501038Z",
                    "step_name": "Pre-Processing",
                    "step_start_time": "2016-03-27T10:22:36.4500059Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:46.4504326Z",
                    "step_name": "Processing",
                    "step_start_time": "2016-03-27T10:22:41.4501038Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:51.4504397Z",
                    "step_name": "Post-Processing",
                    "step_start_time": "2016-03-27T10:22:46.4504326Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:51.4504397Z",
                    "step_name": "Completed",
                    "step_start_time": "2016-03-27T10:22:51.4504397Z"
                }
            ]

