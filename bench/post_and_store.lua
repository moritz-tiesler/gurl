local shortened_urls = {}
local url_counter = 0
local threads = {}

function setup(thread)
    table.insert(threads, thread)
end

function init(args)
    urls = {}
end


request = function()
    url_counter = url_counter + 1
    local original_url = "https://example.com/unique_url_" .. url_counter .. "_" .. math.random(1, 10000)

    local body = "long_url=" .. original_url
    wrk.method = "POST"
    wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
    wrk.body = body

    return wrk.format()
end


response = function(status, headers, body)
   if status == 200 or status == 201 then -- Assuming your server returns 200 or 201 on success
      local shortened_code = string.match(body, "url/(%S+)") -- Basic example: captures the last part of the URL

      if shortened_code and string.len(shortened_code) > 0 then
            table.insert(urls, shortened_code)
      end
   end
end


done = function(summary, latency, requests)
    local fname = "bench/test_urls.text"
    local file = io.open(fname, "w")
    if file then
        for index, thread in ipairs(threads) do
            local urls = thread:get("urls")
            for j, url in ipairs(urls) do
                file:write(tostring(url), "\n")
            end
        end
    file:close()
    else
        print("cannot open file with name=" .. fname)
    end
end