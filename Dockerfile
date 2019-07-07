FROM scratch

COPY pvctl /bin/

ENTRYPOINT ["/bin/pvctl"]