func (r *Reader) ReadMIMEHeader() (MIMEHeader, error) {
    // Avoid lots of small slice allocations later by allocating one
    // large one ahead of time which we'll cut up into smaller
    // slices. If this isn't big enough later, we allocate small ones.
    var strs []string
    hint := r.upcomingHeaderNewlines()
    if hint > 0 {
        strs = make([]string, hint)
    }

    m := make(MIMEHeader, hint)
    for {
        kv, err := r.readContinuedLineSlice()
        if len(kv) == 0 {
            return m, err
        }

        // Key ends at first colon; should not have spaces but
        // they appear in the wild, violating specs, so we
        // remove them if present.
        i := bytes.IndexByte(kv, ':')
        if i < 0 {
            return m, ProtocolError("malformed MIME header line: " + string(kv))
        }
        endKey := i
        for endKey > 0 && kv[endKey-1] == ' ' {
            endKey--
        }
        key := canonicalMIMEHeaderKey(kv[:endKey])

        // As per RFC 7230 field-name is a token, tokens consist of one or more chars.
        // We could return a ProtocolError here, but better to be liberal in what we
        // accept, so if we get an empty key, skip it.
        if key == "" {
            continue
        }

        // Skip initial spaces in value.
        i++ // skip colon
        for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
            i++
        }
        value := string(kv[i:])

        vv := m[key]
        if vv == nil && len(strs) > 0 {
            // More than likely this will be a single-element key.
            // Most headers aren't multi-valued.
            // Set the capacity on strs[0] to 1, so any future append
            // won't extend the slice into the other strings.
            vv, strs = strs[:1:1], strs[1:]
            vv[0] = value
            m[key] = vv
        } else {
            m[key] = append(vv, value)
        }

        if err != nil {
            return m, err
        }
    }
}

